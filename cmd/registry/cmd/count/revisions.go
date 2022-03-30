// Copyright 2022 Google LLC. All Rights Reserved.
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

package count

import (
	"context"
	"fmt"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/protobuf/field_mask"
)

func revisionsCommand(ctx context.Context) *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "revisions",
		Short: "Count the number of revisions of specified resources",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, 64)
			defer wait()
			// Generate tasks.
			name := args[0]
			if spec, err := names.ParseSpec(name); err == nil {
				err = core.ListSpecs(ctx, client, spec, filter, func(spec *rpc.ApiSpec) error {
					taskQueue <- &countSpecRevisionsTask{
						client: client,
						spec:   spec,
					}
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to list API specs")
				}
			} else if deployment, err := names.ParseDeployment(name); err == nil {
				err = core.ListDeployments(ctx, client, deployment, filter, func(deployment *rpc.ApiDeployment) error {
					taskQueue <- &countDeploymentRevisionsTask{
						client:     client,
						deployment: deployment,
					}
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to list API deployments")
				}
			} else {
				log.FromContext(ctx).WithError(err).Fatal("Unsupported resource")
			}
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}

type countSpecRevisionsTask struct {
	client connection.Client
	spec   *rpc.ApiSpec
}

func (task *countSpecRevisionsTask) String() string {
	return "count revisions " + task.spec.Name
}

func (task *countSpecRevisionsTask) Run(ctx context.Context) error {
	name, err := names.ParseSpec(task.spec.Name)
	if err != nil {
		return err
	}
	count := 0
	err = core.ListSpecRevisions(ctx, task.client, name, "", func(*rpc.ApiSpec) error {
		count++
		return nil
	})
	if err != nil {
		return err
	}
	log.Debugf(ctx, "%d\t%s", count, task.spec.Name)
	if task.spec.Labels == nil {
		task.spec.Labels = make(map[string]string, 0)
	}
	task.spec.Labels["revisions"] = fmt.Sprintf("%d", count)
	_, err = task.client.UpdateApiSpec(ctx,
		&rpc.UpdateApiSpecRequest{
			ApiSpec: task.spec,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}

type countDeploymentRevisionsTask struct {
	client     connection.Client
	deployment *rpc.ApiDeployment
}

func (task *countDeploymentRevisionsTask) String() string {
	return "count revisions " + task.deployment.Name
}

func (task *countDeploymentRevisionsTask) Run(ctx context.Context) error {
	count := 0
	it := task.client.ListApiDeploymentRevisions(ctx,
		&rpc.ListApiDeploymentRevisionsRequest{
			Name: task.deployment.Name,
		})
	for {
		_, err := it.Next()
		if err == iterator.Done {
			break
		} else if err == nil {
			count++
		} else {
			return err
		}
	}
	log.Debugf(ctx, "%-7d %s", count, task.deployment.Name)
	if task.deployment.Labels == nil {
		task.deployment.Labels = make(map[string]string, 0)
	}
	task.deployment.Labels["revisions"] = fmt.Sprintf("%d", count)
	_, err := task.client.UpdateApiDeployment(ctx,
		&rpc.UpdateApiDeploymentRequest{
			ApiDeployment: task.deployment,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}
