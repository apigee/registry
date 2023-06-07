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

package compute

import (
	"github.com/apigee/registry/cmd/registry/cmd/compute/complexity"
	"github.com/apigee/registry/cmd/registry/cmd/compute/conformance"
	"github.com/apigee/registry/cmd/registry/cmd/compute/lint"
	"github.com/apigee/registry/cmd/registry/cmd/compute/lintstats"
	"github.com/apigee/registry/cmd/registry/cmd/compute/score"
	"github.com/apigee/registry/cmd/registry/cmd/compute/scorecard"
	"github.com/apigee/registry/cmd/registry/cmd/compute/vocabulary"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Compute properties of resources in the API Registry",
	}

	cmd.AddCommand(conformance.Command())
	cmd.AddCommand(complexity.Command())
	cmd.AddCommand(lint.Command())
	cmd.AddCommand(lintstats.Command())
	cmd.AddCommand(score.Command())
	cmd.AddCommand(scorecard.Command())
	cmd.AddCommand(vocabulary.Command())

	return cmd
}
