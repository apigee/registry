// Copyright 2021 Google LLC. All Rights Reserved.
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

package annotate

import (
	"context"
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/genproto/protobuf/field_mask"
)

func Command(ctx context.Context) *cobra.Command {
	var (
		filter    string
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:   "annotate RESOURCE KEY_1=VAL_1 ... KEY_N=VAL_N",
		Short: "Annotate resources in the API Registry",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}

			taskQueue, wait := core.WorkerPool(ctx, 64)
			defer wait()

			valuesToClear := make([]string, 0)
			valuesToSet := make(map[string]string)
			for _, operation := range args[1:] {
				if len(operation) > 1 && strings.HasSuffix(operation, "-") {
					valuesToClear = append(valuesToClear, strings.TrimSuffix(operation, "-"))
				} else {
					pair := strings.Split(operation, "=")
					if len(pair) != 2 {
						log.Fatalf("%q must have the form \"key=value\" (value can be empty) or \"key-\" (to remove the key)", operation)
					}
					if pair[0] == "" {
						log.Fatalf("%q is invalid because it specifies an empty key", operation)
					}
					valuesToSet[pair[0]] = pair[1]
				}
			}
			labeling := &core.Labeling{Overwrite: overwrite, Set: valuesToSet, Clear: valuesToClear}

			err = matchAndHandleAnnotateCmd(ctx, client, taskQueue, args[0], filter, labeling)
			if err != nil {
				log.WithError(err).Fatal("Failed to handle command")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing annotations")
	return cmd
}

func matchAndHandleAnnotateCmd(
	ctx context.Context,
	client connection.Client,
	taskQueue chan<- core.Task,
	name string,
	filter string,
	labeling *core.Labeling,
) error {
	// First try to match collection names.
	if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		return annotateAPIs(ctx, client, m, filter, labeling, taskQueue)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		return annotateVersions(ctx, client, m, filter, labeling, taskQueue)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		return annotateSpecs(ctx, client, m, filter, labeling, taskQueue)
	}

	// Then try to match resource names.
	if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return annotateAPIs(ctx, client, m, filter, labeling, taskQueue)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return annotateVersions(ctx, client, m, filter, labeling, taskQueue)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return annotateSpecs(ctx, client, m, filter, labeling, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func annotateAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, segments, filterFlag, func(api *rpc.Api) {
		taskQueue <- &annotateApiTask{
			client:   client,
			api:      api,
			labeling: labeling,
		}
	})
}

func annotateVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, segments, filterFlag, func(version *rpc.ApiVersion) {
		taskQueue <- &annotateVersionTask{
			client:   client,
			version:  version,
			labeling: labeling,
		}
	})
}

func annotateSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, segments, filterFlag, func(spec *rpc.ApiSpec) {
		taskQueue <- &annotateSpecTask{
			client:   client,
			spec:     spec,
			labeling: labeling,
		}
	})
}

type annotateApiTask struct {
	client   connection.Client
	api      *rpc.Api
	labeling *core.Labeling
}

func (task *annotateApiTask) String() string {
	return "annotate " + task.api.Name
}

func (task *annotateApiTask) Run(ctx context.Context) error {
	var err error
	task.api.Annotations, err = task.labeling.Apply(task.api.Annotations)
	if err != nil {
		log.WithError(err).Errorf("Invalid annotation")
		return nil
	}
	_, err = task.client.UpdateApi(ctx,
		&rpc.UpdateApiRequest{
			Api: task.api,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"annotations"},
			},
		})
	return err
}

type annotateVersionTask struct {
	client   connection.Client
	version  *rpc.ApiVersion
	labeling *core.Labeling
}

func (task *annotateVersionTask) String() string {
	return "annotate " + task.version.Name
}

func (task *annotateVersionTask) Run(ctx context.Context) error {
	var err error
	task.version.Annotations, err = task.labeling.Apply(task.version.Annotations)
	if err != nil {
		log.WithError(err).Errorf("Invalid annotation")
		return nil
	}
	_, err = task.client.UpdateApiVersion(ctx,
		&rpc.UpdateApiVersionRequest{
			ApiVersion: task.version,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"annotations"},
			},
		})
	return err
}

type annotateSpecTask struct {
	client   connection.Client
	spec     *rpc.ApiSpec
	labeling *core.Labeling
}

func (task *annotateSpecTask) String() string {
	return "annotate " + task.spec.Name
}

func (task *annotateSpecTask) Run(ctx context.Context) error {
	var err error
	task.spec.Annotations, err = task.labeling.Apply(task.spec.Annotations)
	if err != nil {
		log.WithError(err).Errorf("Invalid annotation")
		return nil
	}
	_, err = task.client.UpdateApiSpec(ctx,
		&rpc.UpdateApiSpecRequest{
			ApiSpec: task.spec,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"annotations"},
			},
		})
	return err
}
