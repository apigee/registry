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

package check

import (
	"log"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	globalRules   = lint.NewRuleRegistry()
	globalConfigs = defaultConfigs()
)

func init() {
	if err := rules.Add(globalRules); err != nil {
		log.Fatalf("error when registering rules: %v", err)
	}
}

// Enable all rules by default.
func defaultConfigs() lint.Configs {
	return lint.Configs{}
}

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [pattern]",
		Short: "Check entities in the API Registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}

			root, err := names.Parse(c.FQName(args[0]))
			if err != nil {
				return err
			}

			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				return err
			}

			jobs, err := cmd.Flags().GetInt("jobs")
			if err != nil {
				return err
			}

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				return err
			}

			linter := lint.New(globalRules, globalConfigs)
			response, err := linter.Check(ctx, adminClient, client, root, filter, jobs)
			if err != nil {
				return err
			}

			serialized, err := yaml.Marshal(response)
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(serialized)
			return err
		},
	}

	cmd.PersistentFlags().String("filter", "", "Filter selected resources")
	cmd.PersistentFlags().Int("jobs", 10, "Number of actions to perform concurrently")

	return cmd
}
