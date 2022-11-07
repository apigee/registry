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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
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
		Use:   "annotate RESOURCE KEY_1=VAL_1 ... KEY_N=VAL_N",
		Short: "Annotate resources in the API Registry",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			args[0] = c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			taskQueue, wait := core.WorkerPool(ctx, jobs)
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

			err = matchAndHandleAnnotateCmd(ctx, client, taskQueue, args[0], filter, labeling)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to handle command")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing annotations")
	cmd.Flags().IntVar(&jobs, "jobs", 10, "Number of actions to perform concurrently")
	return cmd
}

func matchAndHandleAnnotateCmd(
	ctx context.Context,
	client connection.RegistryClient,
	taskQueue chan<- core.Task,
	name string,
	filter string,
	labeling *core.Labeling,
) error {
	// First try to match collection names.
	if api, err := names.ParseApiCollection(name); err == nil {
		return annotateAPIs(ctx, client, api, filter, labeling, taskQueue)
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return annotateVersions(ctx, client, version, filter, labeling, taskQueue)
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return annotateSpecs(ctx, client, spec, filter, labeling, taskQueue)
	} else if deployment, err := names.ParseDeploymentCollection(name); err == nil {
		return annotateDeployments(ctx, client, deployment, filter, labeling, taskQueue)
	}

	// Then try to match resource names.
	if api, err := names.ParseApi(name); err == nil {
		return annotateAPIs(ctx, client, api, filter, labeling, taskQueue)
	} else if version, err := names.ParseVersion(name); err == nil {
		return annotateVersions(ctx, client, version, filter, labeling, taskQueue)
	} else if spec, err := names.ParseSpec(name); err == nil {
		return annotateSpecs(ctx, client, spec, filter, labeling, taskQueue)
	} else if deployment, err := names.ParseDeployment(name); err == nil {
		return annotateDeployments(ctx, client, deployment, filter, labeling, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func annotateAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	api names.Api,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, api, filterFlag, func(api *rpc.Api) error {
		taskQueue <- &annotateApiTask{
			client:   client,
			api:      api,
			labeling: labeling,
		}
		return nil
	})
}

func annotateVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	version names.Version,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, version, filterFlag, func(version *rpc.ApiVersion) error {
		taskQueue <- &annotateVersionTask{
			client:   client,
			version:  version,
			labeling: labeling,
		}
		return nil
	})
}

func annotateSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	spec names.Spec,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, spec, filterFlag, func(spec *rpc.ApiSpec) error {
		taskQueue <- &annotateSpecTask{
			client:   client,
			spec:     spec,
			labeling: labeling,
		}
		return nil
	})
}

func annotateDeployments(
	ctx context.Context,
	client *gapic.RegistryClient,
	deployment names.Deployment,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListDeployments(ctx, client, deployment, filterFlag, func(deployment *rpc.ApiDeployment) error {
		taskQueue <- &annotateDeploymentTask{
			client:     client,
			deployment: deployment,
			labeling:   labeling,
		}
		return nil
	})
}

type annotateApiTask struct {
	client   connection.RegistryClient
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
		log.FromContext(ctx).WithError(err).Errorf("Invalid annotation")
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
	client   connection.RegistryClient
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
		log.FromContext(ctx).WithError(err).Errorf("Invalid annotation")
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
	client   connection.RegistryClient
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
		log.FromContext(ctx).WithError(err).Errorf("Invalid annotation")
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

type annotateDeploymentTask struct {
	client     connection.RegistryClient
	deployment *rpc.ApiDeployment
	labeling   *core.Labeling
}

func (task *annotateDeploymentTask) String() string {
	return "annotate " + task.deployment.Name
}

func (task *annotateDeploymentTask) Run(ctx context.Context) error {
	var err error
	task.deployment.Annotations, err = task.labeling.Apply(task.deployment.Annotations)
	if err != nil {
		log.FromContext(ctx).WithError(err).Errorf("Invalid annotation")
		return nil
	}
	_, err = task.client.UpdateApiDeployment(ctx,
		&rpc.UpdateApiDeploymentRequest{
			ApiDeployment: task.deployment,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"annotations"},
			},
		})
	return err
}
