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

package bulk

import (
	"testing"
)

func TestDiscoveryMissingParent(t *testing.T) {
	const (
		projectID   = "missing"
		projectName = "projects/" + projectID
		parent      = projectName + "/locations/global"
	)
	tests := []struct {
		desc string
		args []string
	}{
		{
			desc: "parent",
			args: []string{"discovery", "--parent", parent},
		},
		{
			desc: "project-id",
			args: []string{"discovery", "--project-id", projectID},
		},
		{
			desc: "unspecified",
			args: []string{"discovery"},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs(test.args)
			if cmd.Execute() == nil {
				t.Error("expected error, none reported")
			}
		})
	}
}
