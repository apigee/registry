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
	"encoding/hex"
	"fmt"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"log"
	"os"
	"os/exec"
	"strings"
)

type ExecCommandTask struct {
	Action            string
	TaskID            string
	Placeholder       bool
	GeneratedResource string
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
	if task.Placeholder {
		err := touchArtifact(ctx, task.GeneratedResource, task.Action)
		if err != nil {
			log.Printf("Failed Execution: %s Error: Failed updating placeholder %s", taskDetails, err)
		}
	}
	log.Printf("Successful Execution: %s", taskDetails)
	return nil
}

func touchArtifact(
	ctx context.Context,
	artifactName string,
	action string,
) error {
	client, err := connection.NewClient(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	// TODO: Define a proto for storing placeholder artifacts
	contents := []byte(action)
	encodedContents := make([]byte, hex.EncodedLen(len(contents)))
	hex.Encode(encodedContents, contents)

	artifact := &rpc.Artifact{
		Name:     artifactName,
		Contents: encodedContents,
	}

	err = core.SetArtifact(ctx, client, artifact)
	if err != nil {
		return err
	}
	return nil
}
