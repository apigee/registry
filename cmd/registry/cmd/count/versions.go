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

func versionsCommand() *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "Count the number of versions of specified APIs",
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

			api, err := names.ParseApi(args[0])
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed parse")
			}

			// Iterate through a collection of APIs and count the number of versions of each.
			err = core.ListAPIs(ctx, client, api, filter, func(api *rpc.Api) error {
				taskQueue <- &countApiVersionsTask{
					client: client,
					api:    api,
				}
				return nil
			})
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list APIs")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	return cmd
}

type countApiVersionsTask struct {
	client connection.RegistryClient
	api    *rpc.Api
}

func (task *countApiVersionsTask) String() string {
	return "count versions " + task.api.Name
}

func (task *countApiVersionsTask) Run(ctx context.Context) error {
	count := 0
	request := &rpc.ListApiVersionsRequest{
		Parent: task.api.Name,
	}
	it := task.client.ListApiVersions(ctx, request)
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
	log.Debugf(ctx, "%d\t%s", count, task.api.Name)
	if task.api.Labels == nil {
		task.api.Labels = make(map[string]string, 0)
	}
	task.api.Labels["versions"] = fmt.Sprintf("%d", count)
	_, err := task.client.UpdateApi(ctx,
		&rpc.UpdateApiRequest{
			Api: task.api,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"labels"},
			},
		})
	return err
}
