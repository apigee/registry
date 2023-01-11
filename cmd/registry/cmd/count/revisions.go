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
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/protobuf/field_mask"
)

func revisionsCommand() *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "revisions",
		Short: "Count the number of revisions of specified resources",
		Args:  cobra.ExactArgs(1),
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
			// Initialize task queue.
			jobs, err := cmd.Flags().GetInt("jobs")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get jobs from flags")
			}
			taskQueue, wait := core.WorkerPool(ctx, jobs)
			defer wait()
			// Generate tasks.
			if spec, err := names.ParseSpec(args[0]); err == nil {
				err = core.ListSpecs(ctx, client, spec, filter, false, func(spec *rpc.ApiSpec) error {
					taskQueue <- &countSpecRevisionsTask{
						client:     client,
						specName:   spec.Name,
						specLabels: spec.Labels,
					}
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to list API specs")
				}
			} else if deployment, err := names.ParseDeployment(args[0]); err == nil {
				err = core.ListDeployments(ctx, client, deployment, filter, func(deployment *rpc.ApiDeployment) error {
					taskQueue <- &countDeploymentRevisionsTask{
						client:           client,
						deploymentName:   deployment.Name,
						deploymentLabels: deployment.Labels,
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
	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	return cmd
}

type countSpecRevisionsTask struct {
	client     connection.RegistryClient
	specName   string
	specLabels map[string]string
}

func (task *countSpecRevisionsTask) String() string {
	return "count revisions " + task.specName
}

func (task *countSpecRevisionsTask) Run(ctx context.Context) error {
	name, err := names.ParseSpecRevision(task.specName)
	if err != nil {
		return err
	}
	count := 0
	err = core.ListSpecRevisions(ctx, task.client, name, "", false, func(*rpc.ApiSpec) error {
		count++
		return nil
	})
	if err != nil {
		return err
	}
	log.Debugf(ctx, "%d\t%s", count, task.specName)
	if task.specLabels == nil {
		task.specLabels = make(map[string]string, 0)
	}
	task.specLabels["revisions"] = fmt.Sprintf("%d", count)
	_, err = task.client.UpdateApiSpec(ctx,
		&rpc.UpdateApiSpecRequest{
			ApiSpec: &rpc.ApiSpec{
				Name:   task.specName,
				Labels: task.specLabels,
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}

type countDeploymentRevisionsTask struct {
	client           connection.RegistryClient
	deploymentName   string
	deploymentLabels map[string]string
}

func (task *countDeploymentRevisionsTask) String() string {
	return "count revisions " + task.deploymentName
}

func (task *countDeploymentRevisionsTask) Run(ctx context.Context) error {
	count := 0
	it := task.client.ListApiDeploymentRevisions(ctx,
		&rpc.ListApiDeploymentRevisionsRequest{
			Name: task.deploymentName,
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
	log.Debugf(ctx, "%-7d %s", count, task.deploymentName)
	if task.deploymentLabels == nil {
		task.deploymentLabels = make(map[string]string, 0)
	}
	task.deploymentLabels["revisions"] = fmt.Sprintf("%d", count)
	_, err := task.client.UpdateApiDeployment(ctx,
		&rpc.UpdateApiDeploymentRequest{
			ApiDeployment: &rpc.ApiDeployment{
				Name:   task.deploymentName,
				Labels: task.deploymentLabels,
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}
