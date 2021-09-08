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

package cmd

import (
	"context"

	"github.com/apigee/registry/cmd/registry/cmd/annotate"
	"github.com/apigee/registry/cmd/registry/cmd/compute"
	"github.com/apigee/registry/cmd/registry/cmd/delete"
	"github.com/apigee/registry/cmd/registry/cmd/exec"
	"github.com/apigee/registry/cmd/registry/cmd/export"
	"github.com/apigee/registry/cmd/registry/cmd/get"
	"github.com/apigee/registry/cmd/registry/cmd/index"
	"github.com/apigee/registry/cmd/registry/cmd/label"
	"github.com/apigee/registry/cmd/registry/cmd/list"
	"github.com/apigee/registry/cmd/registry/cmd/resolve"
	"github.com/apigee/registry/cmd/registry/cmd/search"
	"github.com/apigee/registry/cmd/registry/cmd/upload"
	"github.com/apigee/registry/cmd/registry/cmd/vocabulary"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "registry",
		Short: "A simple and eclectic utility for working with the API Registry",
	}

	cmd.AddCommand(annotate.Command(ctx))
	cmd.AddCommand(compute.Command(ctx))
	cmd.AddCommand(resolve.Command(ctx))
	cmd.AddCommand(delete.Command(ctx))
	cmd.AddCommand(exec.Command(ctx))
	cmd.AddCommand(export.Command(ctx))
	cmd.AddCommand(get.Command(ctx))
	cmd.AddCommand(index.Command(ctx))
	cmd.AddCommand(label.Command(ctx))
	cmd.AddCommand(list.Command(ctx))
	cmd.AddCommand(search.Command(ctx))
	cmd.AddCommand(upload.Command(ctx))
	cmd.AddCommand(vocabulary.Command(ctx))

	return cmd
}
