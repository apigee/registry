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

package export

import (
	"encoding/csv"
	"fmt"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

type exportCSVRow struct {
	ApiID        string
	VersionID    string
	SpecID       string
	ContentsPath string
}

func csvCommand() *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "csv [--filter expression] parent ...",
		Short: "Export API specs to a CSV file",
		Args: func(cmd *cobra.Command, args []string) error {
			c, err := connection.ActiveConfig()
			if err != nil {
				return fmt.Errorf("Failed to get config: %s", err)
			}
			for _, parent := range args {
				parent = c.FQName(parent)
				if _, err := names.ParseVersion(parent); err != nil {
					return fmt.Errorf("invalid parent argument %q", parent)
				}
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			rows := make([]exportCSVRow, 0)
			for _, parent := range args {
				version, err := names.ParseVersion(parent)
				if err != nil {
					log.Debugf(ctx, "Failed to parse version name %q: skipping spec row", parent)
					return
				}

				err = core.ListSpecs(ctx, client, version.Spec(""), filter, false, func(spec *rpc.ApiSpec) error {
					name, err := names.ParseSpec(spec.GetName())
					if err != nil {
						log.Debugf(ctx, "Failed to parse spec name %q: skipping spec row", spec.GetName())
						return nil
					}

					rows = append(rows, exportCSVRow{
						ApiID:        name.ApiID,
						VersionID:    name.VersionID,
						SpecID:       name.SpecID,
						ContentsPath: fmt.Sprintf("$APG_REGISTRY_ADDRESS/%s", spec.GetName()),
					})
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatalf("Failed to list specs")
				}
			}

			w := csv.NewWriter(cmd.OutOrStdout())
			if err := w.Write([]string{"api_id", "version_id", "spec_id", "contents_path"}); err != nil {
				log.FromContext(ctx).WithError(err).Fatalf("Failed to write CSV header")
			}

			for _, row := range rows {
				if err := w.Write([]string{row.ApiID, row.VersionID, row.SpecID, row.ContentsPath}); err != nil {
					log.FromContext(ctx).WithError(err).Fatalf("Failed to write CSV row %v", row)
				}
			}

			w.Flush()
			if err := w.Error(); err != nil {
				log.FromContext(ctx).WithError(err).Fatalf("Failed to flush writes to output")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}
