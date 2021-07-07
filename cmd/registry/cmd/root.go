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
	"github.com/apigee/registry/cmd/registry/cmd/annotate"
	"github.com/apigee/registry/cmd/registry/cmd/compute"
	"github.com/apigee/registry/cmd/registry/cmd/controller"
	"github.com/apigee/registry/cmd/registry/cmd/delete"
	"github.com/apigee/registry/cmd/registry/cmd/export"
	"github.com/apigee/registry/cmd/registry/cmd/get"
	"github.com/apigee/registry/cmd/registry/cmd/index"
	"github.com/apigee/registry/cmd/registry/cmd/label"
	"github.com/apigee/registry/cmd/registry/cmd/list"
	"github.com/apigee/registry/cmd/registry/cmd/search"
	"github.com/apigee/registry/cmd/registry/cmd/upload"
	"github.com/apigee/registry/cmd/registry/cmd/vocabulary"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "registry",
		Short: "A simple and eclectic utility for working with the API Registry",
	}

	cmd.AddCommand(annotate.Command())
	cmd.AddCommand(compute.Command())
	cmd.AddCommand(controller.Command())
	cmd.AddCommand(delete.Command())
	cmd.AddCommand(export.Command())
	cmd.AddCommand(get.Command())
	cmd.AddCommand(index.Command())
	cmd.AddCommand(label.Command())
	cmd.AddCommand(list.Command())
	cmd.AddCommand(search.Command())
	cmd.AddCommand(upload.Command())
	cmd.AddCommand(vocabulary.Command())

	return cmd
}
