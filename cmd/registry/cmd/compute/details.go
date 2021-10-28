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

package compute

import (
	"context"
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/genproto/protobuf/field_mask"

	discovery "github.com/googleapis/gnostic/discovery"
	oas2 "github.com/googleapis/gnostic/openapiv2"
	oas3 "github.com/googleapis/gnostic/openapiv3"
)

func detailsCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "details",
		Short: "Compute details about APIs from information in their specs.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.WithError(err).Fatal("Failed to get filter from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}
			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, 64)
			defer wait()
			// Generate tasks.
			name := args[0]
			if api, err := names.ParseApi(name); err == nil {
				// Iterate through a collection of specs and summarize each.
				err = core.ListAPIs(ctx, client, api, filter, func(api *rpc.Api) {
					taskQueue <- &computeDetailsTask{
						client:  client,
						apiName: api.Name,
					}
				})
				if err != nil {
					log.WithError(err).Debugf("Failed to list APIs")
				}
			}
		},
	}
}

type computeDetailsTask struct {
	client  connection.Client
	apiName string
}

func (task *computeDetailsTask) String() string {
	return "compute details " + task.apiName
}

func (task *computeDetailsTask) Run(ctx context.Context) error {
	specName, err := names.ParseSpec(task.apiName + "/versions/-/specs/-")
	if err != nil {
		return err
	}
	specs := make([]*rpc.ApiSpec, 0)
	_ = core.ListSpecs(ctx, task.client, specName, "", func(spec *rpc.ApiSpec) {
		specs = append(specs, spec)
	})
	// use the last (presumed latest) spec
	if len(specs) == 0 {
		return nil
	}
	spec := specs[len(specs)-1]
	specName, err = names.ParseSpec(spec.Name)
	if err != nil {
		return nil
	}
	spec, err = core.GetSpec(ctx, task.client, specName, true, nil)
	if err != nil {
		return nil
	}
	var title string
	var description string
	var request *rpc.UpdateApiRequest
	if core.IsOpenAPIv2(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := oas2.ParseDocument(data)
		if document == nil && err != nil {
			log.WithError(err).Warnf("Invalid OpenAPI: %s", spec.Name)
			return nil
		}
		if document.Info != nil {
			title = document.Info.Title
			description = document.Info.Description
		}
		if len(description) > 256 {
			description = description[0:256]
		}
		request = &rpc.UpdateApiRequest{
			Api: &rpc.Api{
				Name:        task.apiName,
				DisplayName: title,
				Description: description,
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"display_name", "description"},
			},
		}
	} else if core.IsOpenAPIv3(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := oas3.ParseDocument(data)
		if document == nil && err != nil {
			log.WithError(err).Warnf("Invalid OpenAPI: %s", spec.Name)
			return nil
		}
		if document.Info != nil {
			title = document.Info.Title
			description = document.Info.Description
		}
		if len(description) > 256 {
			description = description[0:256]
		}
		request = &rpc.UpdateApiRequest{
			Api: &rpc.Api{
				Name:        task.apiName,
				DisplayName: title,
				Description: description,
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"display_name", "description"},
			},
		}
	} else if core.IsDiscovery(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := discovery.ParseDocument(data)
		if document == nil && err != nil {
			log.WithError(err).Warnf("Invalid Discovery document: %s", spec.Name)
			return nil
		}
		title := document.Title
		description := document.Description
		if len(description) > 256 {
			description = description[0:256]
		}
		request = &rpc.UpdateApiRequest{
			Api: &rpc.Api{
				Name:        task.apiName,
				DisplayName: title,
				Description: description,
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"display_name", "description"},
			},
		}

	} else if core.IsProto(spec.GetMimeType()) && core.IsZipArchive(spec.GetMimeType()) {
		log.Debug(spec.Name)
		details, err := core.NewDetailsFromZippedProtos(spec.GetContents())
		if err != nil {
			log.WithError(err).Warnf("Error processing protos: %s", spec.Name)
			return nil
		}
		if details != nil {
			title := details.Title
			var description string
			if len(details.Services) == 0 {
				description = "0 Services"
			} else if len(details.Services) == 1 {
				description = fmt.Sprintf("1 Service: %s", details.Services[0])
			} else {
				description = fmt.Sprintf("%d Services: ", len(details.Services)) + strings.Join(details.Services, ", ")
			}
			if len(description) > 256 {
				description = description[0:256] + "..."
			}
			request = &rpc.UpdateApiRequest{
				Api: &rpc.Api{
					Name:        task.apiName,
					DisplayName: title,
					Description: description,
				},
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{"display_name", "description"},
				},
			}
		}
	} else {
		log.Warnf("We don't know how to compute the title of %s", task.apiName)
		return nil
	}
	if request != nil {
		_, err = task.client.UpdateApi(ctx, request)
		if err != nil {
			log.WithError(err).Warnf("Error updating API: %s", task.apiName)
			return nil
		}
	}
	return err
}
