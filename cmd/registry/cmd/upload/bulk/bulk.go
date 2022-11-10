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

package bulk

import (
	"errors"
	"fmt"

	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk-upload API specs of selected styles",
	}

	cmd.AddCommand(discoveryCommand())
	cmd.AddCommand(openAPICommand())
	cmd.AddCommand(protosCommand())

	cmd.PersistentFlags().String("project-id", "", "Project ID to use for each upload (deprecated)")
	cmd.PersistentFlags().String("parent", "", "Parent for the upload (projects/PROJECT/locations/LOCATION)")
	cmd.PersistentFlags().Int("jobs", 10, "Number of upload jobs to run simultaneously")
	return cmd
}

func getParent(cmd *cobra.Command) (string, error) {
	ctx := cmd.Context()

	parent, err := cmd.Flags().GetString("parent")
	if err != nil {
		return "", fmt.Errorf("failed to get parent from flags (%s)", err)
	}
	projectID, err := cmd.Flags().GetString("project-id")
	if err != nil {
		return "", fmt.Errorf("failed to get project-id from flags (%s)", err)
	}
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
