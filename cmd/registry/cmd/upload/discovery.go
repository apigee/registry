// Copyright 2020 Google LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upload

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"

	"github.com/google/gnostic/compiler"
	discovery "github.com/google/gnostic/discovery"
)

// Read the list of APIs from the Discovery service.
func fetchDiscoveryList(service string) (*discovery.List, error) {
	if service == "" {
		service = discovery.APIsListServiceURL
	}
	bytes, err := compiler.FetchFile(service)
	if err != nil {
		return nil, err
	}
	return discovery.ParseList(bytes)
}

func discoveryCommand() *cobra.Command {
	var service string
	var jobs int
	cmd := &cobra.Command{
		Use:   "discovery",
		Args:  cobra.NoArgs,
		Short: "Upload API Discovery documents from the Google API Discovery service",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			parent, err := getParent(cmd)
			if err != nil {
				return fmt.Errorf("failed to identify parent project (%s)", err)
			}
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				return err
			}
			if err := visitor.VerifyLocation(ctx, client, parent); err != nil {
				return fmt.Errorf("parent does not exist (%s)", err)
			}
			// create a queue for upload tasks and wait for the workers to finish after filling it.
			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()

			discoveryResponse, err := fetchDiscoveryList(service)
			if err != nil {
				return fmt.Errorf("failed to fetch discovery list: %s", err)
			}

			// Create an upload job for each API.
			for _, api := range discoveryResponse.APIs {
				taskQueue <- &uploadDiscoveryTask{
					client:    client,
					path:      api.DiscoveryRestURL,
					parent:    parent,
					apiID:     sanitize(api.Name),
					versionID: sanitize(api.Version),
					specID:    "discovery",
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&projectID, "project-id", "", "project ID to use for each upload (deprecated)")
	cmd.Flags().StringVar(&parent, "parent", "", "parent for the upload (projects/PROJECT/locations/LOCATION)")
	cmd.Flags().StringVar(&service, "service", "",
		fmt.Sprintf("the API Discovery Service URL (default %s)", discovery.APIsListServiceURL))
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
}

type uploadDiscoveryTask struct {
	client    connection.RegistryClient
	path      string
	parent    string
	apiID     string
	versionID string
	specID    string
	contents  []byte
	info      DiscoveryInfo
}

func (task *uploadDiscoveryTask) String() string {
	return "upload discovery " + task.path
}

func (task *uploadDiscoveryTask) Run(ctx context.Context) error {
	log.Infof(ctx, "Uploading apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)
	// Fetch the contents of the discovery doc.
	// Do this first in case the doc URL is invalid; we skip APIs with these errors.
	if err := task.fetchDiscoveryDoc(); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to download discovery doc")
		return nil
	}
	// If the API does not exist, create it.
	if err := task.createAPI(ctx); err != nil {
		return err
	}
	// If the API version does not exist, create it.
	if err := task.createVersion(ctx); err != nil {
		return err
	}
	// Create or update the spec as needed.
	if err := task.createOrUpdateSpec(ctx); err != nil {
		return err
	}
	return nil
}

func (task *uploadDiscoveryTask) createAPI(ctx context.Context) error {
	// Create an API if needed (or update an existing one)
	response, err := task.client.UpdateApi(ctx, &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:        task.apiName(),
			DisplayName: task.info.Title,
			Description: task.info.Description,
		},
		AllowMissing: true,
	})
	if err == nil {
		log.Debugf(ctx, "Updated %s", response.Name)
	} else if status.Code(err) == codes.AlreadyExists {
		log.Debugf(ctx, "Found %s", task.apiName())
	} else {
		log.FromContext(ctx).WithError(err).Debugf("Failed to create API %s", task.apiName())
		// Returning this error ends all tasks, which seems appropriate to
		// handle situations where all might fail due to a common problem
		// (a missing project or incorrect project-id).
		return fmt.Errorf("failed to create %s, %s", task.apiName(), err)
	}

	return nil
}

func (task *uploadDiscoveryTask) createVersion(ctx context.Context) error {
	// Create an API version if needed (or update an existing one)
	response, err := task.client.UpdateApiVersion(ctx, &rpc.UpdateApiVersionRequest{
		ApiVersion: &rpc.ApiVersion{
			Name: task.versionName(),
		},
		AllowMissing: true,
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Debugf("Failed to create version %s", task.versionName())
	} else {
		log.Debugf(ctx, "Updated %s", response.Name)
	}

	return nil
}

func (task *uploadDiscoveryTask) createOrUpdateSpec(ctx context.Context) error {
	// Use the spec size and hash to avoid unnecessary uploads.
	spec, err := task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	})

	if err == nil && int(spec.GetSizeBytes()) == len(task.contents) && spec.GetHash() == hashForBytes(task.contents) {
		log.Debugf(ctx, "Matched already uploaded spec %s", task.specName())
		return nil
	}

	gzippedContents, err := compress.GZippedBytes(task.contents)
	if err != nil {
		return err
	}

	request := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:      task.specName(),
			MimeType:  mime.DiscoveryMimeType("+gzip"),
			Filename:  "discovery.json",
			Contents:  gzippedContents,
			SourceUri: task.path,
		},
		AllowMissing: true,
	}

	response, err := task.client.UpdateApiSpec(ctx, request)
	if err != nil {
		log.FromContext(ctx).WithError(err).Debugf("Error %s [contents-length: %d]", task.specName(), len(task.contents))
	} else {
		log.Debugf(ctx, "Updated %s", response.Name)
	}

	return nil
}

func (task *uploadDiscoveryTask) apiName() string {
	return fmt.Sprintf("%s/apis/%s", task.parent, task.apiID)
}

func (task *uploadDiscoveryTask) versionName() string {
	return fmt.Sprintf("%s/versions/%s", task.apiName(), task.versionID)
}

func (task *uploadDiscoveryTask) specName() string {
	return fmt.Sprintf("%s/specs/%s", task.versionName(), task.specID)
}

func (task *uploadDiscoveryTask) fetchDiscoveryDoc() error {
	bytes, err := discovery.FetchDocumentBytes(task.path)
	if err != nil {
		return err
	}

	task.contents, err = normalizeJSON(bytes)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(task.contents, &task.info)
}

// Normalize JSON documents by remarshalling them to
// ensure that their keys are sorted alphabetically.
// For readability, marshalled docs are indented.
func normalizeJSON(bytes []byte) ([]byte, error) {
	var m interface{}
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}
	return json.MarshalIndent(m, "", "  ")
}

// A subset of the discovery document useful for adding an API to the registry
type DiscoveryInfo struct {
	Name        string `yaml:"name"`
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}
