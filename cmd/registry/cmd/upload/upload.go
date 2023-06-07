// Copyright 2020 Google LLC.
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

package upload

import (
	"errors"
	"fmt"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload information to the API Registry",
	}

	cmd.AddCommand(csvCommand())
	cmd.AddCommand(discoveryCommand())
	cmd.AddCommand(openAPICommand())
	cmd.AddCommand(protosCommand())
	return cmd
}

// shared among all subcommands
var parent string
var projectID string

func getParent(cmd *cobra.Command) (string, error) {
	ctx := cmd.Context()

	if projectID != "" && parent != "" {
		return "", errors.New("--project-id cannot be used with --parent")
	}
	if parent != "" {
		return parent, nil
	} else if projectID != "" {
		log.FromContext(ctx).Warn("--project-id is deprecated, please use --parent or configure registry.project")
		return "projects/" + projectID + "/locations/global", nil
	}
	c, err := connection.ActiveConfig()
	if err != nil {
		return "", fmt.Errorf("unable to identify parent (%s)", err)
	}
	parent, err = c.ProjectWithLocation()
	if err != nil {
		return "", fmt.Errorf("unable to identify parent: please use --parent or set registry.project in configuration (%s)", err)
	}
	return parent, nil
}
