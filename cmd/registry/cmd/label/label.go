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

package label

import (
	"context"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
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
		Use:   "label RESOURCE KEY_1=VAL_1 ... KEY_N=VAL_N",
		Short: "Label resources in the API Registry",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
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
						log.Fatalf(ctx, "%q must have the form \"key=value\" (value can be empty) or \"key-\" (to remove the key)", operation)
					}
					if pair[0] == "" {
						log.Fatalf(ctx, "%q is invalid because it specifies an empty key", operation)
					}
					valuesToSet[pair[0]] = pair[1]
				}
			}
			labeling := &core.Labeling{Overwrite: overwrite, Set: valuesToSet, Clear: valuesToClear}

			err = matchAndHandleLabelCmd(ctx, client, taskQueue, args[0], filter, labeling)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to match or handle command")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing labels")
	return cmd
}

func matchAndHandleLabelCmd(
	ctx context.Context,
	client connection.Client,
	taskQueue chan<- core.Task,
	name string,
	filter string,
	labeling *core.Labeling,
) error {
	// First try to match collection names.
	if api, err := names.ParseApiCollection(name); err == nil {
		return labelAPIs(ctx, client, api, filter, labeling, taskQueue)
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return labelVersions(ctx, client, version, filter, labeling, taskQueue)
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return labelSpecs(ctx, client, spec, filter, labeling, taskQueue)
	}

	// Then try to match resource names.
	if api, err := names.ParseApi(name); err == nil {
		return labelAPIs(ctx, client, api, filter, labeling, taskQueue)
	} else if version, err := names.ParseVersion(name); err == nil {
		return labelVersions(ctx, client, version, filter, labeling, taskQueue)
	} else if spec, err := names.ParseSpec(name); err == nil {
		return labelSpecs(ctx, client, spec, filter, labeling, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func labelAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	api names.Api,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, api, filterFlag, func(api *rpc.Api) {
		taskQueue <- &labelApiTask{
			client:   client,
			api:      api,
			labeling: labeling,
		}
	})
}

func labelVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	version names.Version,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, version, filterFlag, func(version *rpc.ApiVersion) {
		taskQueue <- &labelVersionTask{
			client:   client,
			version:  version,
			labeling: labeling,
		}
	})
}

func labelSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	spec names.Spec,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, spec, filterFlag, func(spec *rpc.ApiSpec) {
		taskQueue <- &labelSpecTask{
			client:   client,
			spec:     spec,
			labeling: labeling,
		}
	})
}

type labelApiTask struct {
	client   connection.Client
	api      *rpc.Api
	labeling *core.Labeling
}

func (task *labelApiTask) String() string {
	return "label " + task.api.Name
}

func (task *labelApiTask) Run(ctx context.Context) error {
	var err error
	task.api.Labels, err = task.labeling.Apply(task.api.Labels)
	if err != nil {
		log.FromContext(ctx).WithError(err).Errorf("Invalid labelling")
		return nil
	}
	_, err = task.client.UpdateApi(ctx,
		&rpc.UpdateApiRequest{
			Api: task.api,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}

type labelVersionTask struct {
	client   connection.Client
	version  *rpc.ApiVersion
	labeling *core.Labeling
}

func (task *labelVersionTask) String() string {
	return "label " + task.version.Name
}

func (task *labelVersionTask) Run(ctx context.Context) error {
	var err error
	task.version.Labels, err = task.labeling.Apply(task.version.Labels)
	if err != nil {
		log.FromContext(ctx).WithError(err).Errorf("Invalid labelling")
		return nil
	}
	_, err = task.client.UpdateApiVersion(ctx,
		&rpc.UpdateApiVersionRequest{
			ApiVersion: task.version,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}

type labelSpecTask struct {
	client   connection.Client
	spec     *rpc.ApiSpec
	labeling *core.Labeling
}

func (task *labelSpecTask) String() string {
	return "label " + task.spec.Name
}

func (task *labelSpecTask) Run(ctx context.Context) error {
	var err error
	task.spec.Labels, err = task.labeling.Apply(task.spec.Labels)
	if err != nil {
		log.FromContext(ctx).WithError(err).Errorf("Invalid labelling")
		return nil
	}
	_, err = task.client.UpdateApiSpec(ctx,
		&rpc.UpdateApiSpecRequest{
			ApiSpec: task.spec,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}
