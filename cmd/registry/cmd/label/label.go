// Copyright 2021 Google LLC.
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

	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/genproto/protobuf/field_mask"
)

func Command() *cobra.Command {
	var (
		filter    string
		overwrite bool
		jobs      int
	)

	cmd := &cobra.Command{
		Use:   "label RESOURCE KEY_1=VAL_1 ... KEY_N=VAL_N",
		Short: "Label resources in the API Registry",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			args[0] = c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}

			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()

			valuesToClear := make([]string, 0)
			valuesToSet := make(map[string]string)
			for _, operation := range args[1:] {
				if len(operation) > 1 && strings.HasSuffix(operation, "-") {
					valuesToClear = append(valuesToClear, strings.TrimSuffix(operation, "-"))
				} else {
					pair := strings.Split(operation, "=")
					if len(pair) != 2 {
						return fmt.Errorf("%q must have the form \"key=value\" (value can be empty) or \"key-\" (to remove the key)", operation)
					}
					if pair[0] == "" {
						return fmt.Errorf("%q is invalid because it specifies an empty key", operation)
					}
					valuesToSet[pair[0]] = pair[1]
				}
			}
			labeling := &Labeling{Overwrite: overwrite, Set: valuesToSet, Clear: valuesToClear}

			return matchAndHandleLabelCmd(ctx, client, taskQueue, args[0], filter, labeling)
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing labels")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
}

func matchAndHandleLabelCmd(
	ctx context.Context,
	client connection.RegistryClient,
	taskQueue chan<- tasks.Task,
	name string,
	filter string,
	labeling *Labeling,
) error {
	// First try to match collection names.
	if api, err := names.ParseApiCollection(name); err == nil {
		return labelAPIs(ctx, client, api, filter, labeling, taskQueue)
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return labelVersions(ctx, client, version, filter, labeling, taskQueue)
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return labelSpecs(ctx, client, spec, filter, labeling, taskQueue)
	} else if deployment, err := names.ParseDeploymentCollection(name); err == nil {
		return labelDeployments(ctx, client, deployment, filter, labeling, taskQueue)
	}

	// Then try to match resource names.
	if api, err := names.ParseApi(name); err == nil {
		return labelAPIs(ctx, client, api, filter, labeling, taskQueue)
	} else if version, err := names.ParseVersion(name); err == nil {
		return labelVersions(ctx, client, version, filter, labeling, taskQueue)
	} else if spec, err := names.ParseSpec(name); err == nil {
		return labelSpecs(ctx, client, spec, filter, labeling, taskQueue)
	} else if deployment, err := names.ParseDeployment(name); err == nil {
		return labelDeployments(ctx, client, deployment, filter, labeling, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func labelAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	api names.Api,
	filterFlag string,
	labeling *Labeling,
	taskQueue chan<- tasks.Task) error {
	return visitor.ListAPIs(ctx, client, api, 0, filterFlag, func(ctx context.Context, api *rpc.Api) error {
		taskQueue <- &labelApiTask{
			client:   client,
			api:      api,
			labeling: labeling,
		}
		return nil
	})
}

func labelVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	version names.Version,
	filterFlag string,
	labeling *Labeling,
	taskQueue chan<- tasks.Task) error {
	return visitor.ListVersions(ctx, client, version, 0, filterFlag, func(ctx context.Context, version *rpc.ApiVersion) error {
		taskQueue <- &labelVersionTask{
			client:   client,
			version:  version,
			labeling: labeling,
		}
		return nil
	})
}

func labelSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	spec names.Spec,
	filterFlag string,
	labeling *Labeling,
	taskQueue chan<- tasks.Task) error {
	return visitor.ListSpecs(ctx, client, spec, 0, filterFlag, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
		taskQueue <- &labelSpecTask{
			client:   client,
			spec:     spec,
			labeling: labeling,
		}
		return nil
	})
}

func labelDeployments(
	ctx context.Context,
	client *gapic.RegistryClient,
	deployment names.Deployment,
	filterFlag string,
	labeling *Labeling,
	taskQueue chan<- tasks.Task) error {
	return visitor.ListDeployments(ctx, client, deployment, 0, filterFlag, func(ctx context.Context, deployment *rpc.ApiDeployment) error {
		taskQueue <- &labelDeploymentTask{
			client:     client,
			deployment: deployment,
			labeling:   labeling,
		}
		return nil
	})
}

type labelApiTask struct {
	client   connection.RegistryClient
	api      *rpc.Api
	labeling *Labeling
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
	client   connection.RegistryClient
	version  *rpc.ApiVersion
	labeling *Labeling
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
	client   connection.RegistryClient
	spec     *rpc.ApiSpec
	labeling *Labeling
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

type labelDeploymentTask struct {
	client     connection.RegistryClient
	deployment *rpc.ApiDeployment
	labeling   *Labeling
}

func (task *labelDeploymentTask) String() string {
	return "label " + task.deployment.Name
}

func (task *labelDeploymentTask) Run(ctx context.Context) error {
	var err error
	task.deployment.Labels, err = task.labeling.Apply(task.deployment.Labels)
	if err != nil {
		log.FromContext(ctx).WithError(err).Errorf("Invalid labelling")
		return nil
	}
	_, err = task.client.UpdateApiDeployment(ctx,
		&rpc.UpdateApiDeploymentRequest{
			ApiDeployment: task.deployment,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}
