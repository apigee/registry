// Copyright 2022 Google LLC.
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
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
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

var (
	filter     string
	jobs       int
	enable     []string
	disable    []string
	configFile string
	listRules  bool
	errorlevel string
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [PATTERN]",
		Short: "Check entities in the API Registry",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var exitOnErrorLevel check.Problem_Severity
			if errorlevel != "" {
				switch errorlevel {
				case "ERROR":
					exitOnErrorLevel = check.Problem_ERROR
				case "WARNING":
					exitOnErrorLevel = check.Problem_WARNING
				case "INFO":
					exitOnErrorLevel = check.Problem_INFO
				default:
					return fmt.Errorf("invalid level: %q, must be INFO, WARNING, or ERROR", errorlevel)
				}
			}

			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
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

			configs := globalConfigs
			if configFile != "" {
				c, err := lint.ReadConfigsFromFile(configFile)
				if err != nil {
					return err
				}
				configs = append(configs, c...)
			}
			configs = append(configs, lint.Config{
				EnabledRules:  enable,
				DisabledRules: disable,
			})

			if listRules {
				var names []string
				for _, r := range globalRules {
					if configs.IsRuleEnabled(string(r.GetName()), "") {
						names = append(names, string(r.GetName()))
					}
				}
				sort.Strings(names)
				for _, n := range names {
					if _, err := cmd.OutOrStdout().Write([]byte(fmt.Sprintf(" - name: %s\n", n))); err != nil {
						return err
					}
				}
				return nil
			}

			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}
			name := strings.TrimSuffix(c.FQName(args[0]), "/locations/global")
			root, err := names.Parse(name)
			if err != nil {
				return err
			}

			linter := lint.New(globalRules, configs)
			response, err := linter.Check(ctx, adminClient, client, root, filter, jobs)
			if err != nil {
				return err
			}

			serialized, err := yaml.Marshal(response)
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(serialized)
			if err != nil {
				return err
			}

			if exitOnErrorLevel != 0 {
				for _, p := range response.Problems {
					if p.Severity <= exitOnErrorLevel {
						cmd.SilenceUsage = true
						return fmt.Errorf("exceeded designated error-level %q", errorlevel)
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	cmd.Flags().StringVar(&configFile, "config", "", "rule config")
	cmd.Flags().StringArrayVar(&enable, "enable", nil, "enable rules")
	cmd.Flags().StringArrayVar(&disable, "disable", nil, "disable rules")
	cmd.Flags().BoolVar(&listRules, "list-rules", false, "print enabled rules and exit")
	cmd.Flags().StringVar(&errorlevel, "error-level", "", "exit code 1 if problems at specified level or above [INFO|WARNING|ERROR]")

	return cmd
}
