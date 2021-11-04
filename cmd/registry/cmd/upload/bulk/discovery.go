// Copyright 2020 Google LLC. All Rights Reserved.
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

package bulk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"

	discovery "github.com/google/gnostic/discovery"
)

func discoveryCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Bulk-upload API Discovery documents from the Google API Discovery service",
		Run: func(cmd *cobra.Command, args []string) {
			projectID, err := cmd.Flags().GetString("project-id")
			if err != nil {
				log.WithError(err).Fatal("Failed to get project-id from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}

			// create a queue for upload tasks and wait for the workers to finish after filling it.
			taskQueue, wait := core.WorkerPool(ctx, 64)
			defer wait()

			discoveryResponse, err := discovery.FetchList()
			if err != nil {
				log.WithError(err).Fatal("Failed to fetch discovery list")
			}

			// Create an upload job for each API.
			for _, api := range discoveryResponse.APIs {
				taskQueue <- &uploadDiscoveryTask{
					client:    client,
					path:      api.DiscoveryRestURL,
					projectID: projectID,
					apiID:     sanitize(api.Name),
					versionID: sanitize(api.Version),
					specID:    "discovery.json",
				}
			}
		},
	}

	return cmd
}

type uploadDiscoveryTask struct {
	client    connection.Client
	path      string
	projectID string
	apiID     string
	versionID string
	specID    string
	contents  []byte
}

func (task *uploadDiscoveryTask) String() string {
	return "upload discovery " + task.path
}

func (task *uploadDiscoveryTask) Run(ctx context.Context) error {
	log.Infof("Uploading apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)
	// Fetch the contents of the discovery doc.
	// Do this first in case the doc URL is invalid; we skip APIs with these errors.
	if err := task.fetchDiscoveryDoc(); err != nil {
		log.WithError(err).Error("Failed to download discovery doc")
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
			DisplayName: task.apiID,
		},
		AllowMissing: true,
	})
	if err == nil {
		log.Debugf("Created %s", response.Name)
	} else {
		log.WithError(err).Debugf("Failed to create API %s", task.apiName())
		return fmt.Errorf("Failed to create %s, %s", task.apiName(), err)
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
		log.WithError(err).Debugf("Failed to create version %s", task.versionName())
	} else {
		log.Debugf("Created %s", response.Name)
	}

	return nil
}

func (task *uploadDiscoveryTask) createOrUpdateSpec(ctx context.Context) error {
	// Use the spec size and hash to avoid unnecessary uploads.
	if spec, err := task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	}); err == nil {
		if int(spec.GetSizeBytes()) == len(task.contents) {
			hash := hashForBytes(task.contents)
			if spec.GetHash() == hash {
				log.Debugf("Matched existing spec %s", task.specName())
				return nil // this spec is already uploaded
			}
		}
	}

	gzippedContents, err := core.GZippedBytes(task.contents)
	if err != nil {
		return err
	}

	request := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:      task.specName(),
			MimeType:  core.DiscoveryMimeType("+gzip"),
			Filename:  "discovery.json",
			Contents:  gzippedContents,
			SourceUri: task.path,
		},
		AllowMissing: true,
	}

	response, err := task.client.UpdateApiSpec(ctx, request)
	if err != nil {
		log.WithError(err).Debugf("Error %s [contents-length: %d]", task.specName(), len(task.contents))
	} else {
		log.Debugf("Created %s", response.Name)
	}

	return nil
}

func (task *uploadDiscoveryTask) projectName() string {
	return fmt.Sprintf("projects/%s", task.projectID)
}

func (task *uploadDiscoveryTask) apiName() string {
	return fmt.Sprintf("%s/locations/global/apis/%s", task.projectName(), task.apiID)
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

	// normalize the doc to produce consistent hashes
	var m interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}
	normalized, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	task.contents = []byte(normalized)
	return nil
}
