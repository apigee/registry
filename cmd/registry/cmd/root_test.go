// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"github.com/apigee/registry/cmd/registry/controller"
	"testing"
)

// Test that currently supported commands in the registry tool are all covered in ExecCommandTask
func TestCommandCoverage(t *testing.T) {
	ctx := context.Background()
	rootCmd := Command(ctx)

	// Submit tasks to taskQueue
	for i, cmd := range rootCmd.Commands() {
		if cmd.Name() == "resolve" {
			continue
		}
		task := &controller.ExecCommandTask{
			Action: cmd.Name() + " --help",
			TaskID: fmt.Sprintf("task%d", i),
		}
		if err := task.Run(ctx); err != nil {
			t.Errorf("Error executing command: %s, %s", cmd.Name(), err)
		}
	}
}
