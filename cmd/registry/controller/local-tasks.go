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

package controller

import (
	"context"
	"fmt"
	"github.com/apigee/registry/cmd/registry/cmd/annotate"
	"github.com/apigee/registry/cmd/registry/cmd/compute"
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
	"log"
	"strings"
)

type ExecCommandTask struct {
	Action string
	TaskID string
}

func getCommand(ctx context.Context, action []string) (*cobra.Command, error) {
	if len(action) == 0 {
		return nil, fmt.Errorf("Empty action string")
	}

	switch action[0] {
	case "annotate":
		return annotate.Command(ctx), nil
	case "compute":
		return compute.Command(ctx), nil
	case "delete":
		return delete.Command(ctx), nil
	case "export":
		return export.Command(ctx), nil
	case "get":
		return get.Command(ctx), nil
	case "index":
		return index.Command(ctx), nil
	case "label":
		return label.Command(ctx), nil
	case "list":
		return list.Command(ctx), nil
	case "search":
		return search.Command(ctx), nil
	case "upload":
		return upload.Command(ctx), nil
	case "vocabulary":
		return vocabulary.Command(ctx), nil
	default:
		return nil, fmt.Errorf("Action '%s' is not supported.", strings.Join(action, " "))
	}
}

func (task *ExecCommandTask) String() string {
	return "Execute command: " + task.Action
}

func (task *ExecCommandTask) Run(ctx context.Context) error {
	args := strings.Fields(task.Action)
	cmd, err := getCommand(ctx, args)
	if err != nil {
		return err
	}
	// Set args if any
	if len(args) > 1 {
		cmd.SetArgs(args[1:])
	}
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("Error executing action %s: %s", task.Action, err)
	}

	log.Printf("Finished executing taskId %s with action: %s", task.TaskID, task.Action)
	return nil
}
