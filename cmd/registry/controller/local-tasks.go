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
	"log"
	"os"
	"os/exec"
	"strings"
)

type ExecCommandTask struct {
	Action string
	TaskID string
}

func (task *ExecCommandTask) String() string {
	return "Execute command: " + task.Action
}

func (task *ExecCommandTask) Run(ctx context.Context) error {
	//The monitoring metrics/dashboards are built on top of the format of the log messages here.
	//Check the metric filters before making  any changes to the format.
	//Location: registry/deployments/controller/dashboard/*
	taskDetails := fmt.Sprintf("action={%s} taskID={%s}", task.Action, task.TaskID)

	if strings.HasPrefix(task.Action, "resolve") {
		return fmt.Errorf("Failed Execution: %s Error: 'resolve' not allowed in action", taskDetails)
	}
	cmd := exec.Command("registry", strings.Fields(task.Action)...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Printf("Failed Execution: %s Error: %s", taskDetails, err)
		return err
	}
	log.Printf("Successful Execution: %s", taskDetails)
	return nil
}
