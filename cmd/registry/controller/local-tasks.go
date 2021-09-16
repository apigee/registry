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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

type ExecCommandTask struct {
	Action *Action
	TaskID string
}

func (task *ExecCommandTask) String() string {
	return "Execute command: " + task.Action.Command
}

func (task *ExecCommandTask) Run(ctx context.Context) error {
	// The monitoring metrics/dashboards are built on top of the format of the log messages here.
	// Check the metric filters before making any changes to the format.
	// Location: registry/deployments/controller/dashboard/*
	logger := log.WithFields(log.Fields{
		"action": fmt.Sprintf("{%s}", task.Action.Command),
		"taskID": fmt.Sprintf("{%s}", task.TaskID),
	})

	if strings.HasPrefix(task.Action.Command, "resolve") {
		logger.Debug("Failed Execution: 'resolve' not allowed in action")
		return errors.New("'resolve' not allowed in action")
	}

	cmd := exec.Command("registry", strings.Fields(task.Action.Command)...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		logger.WithError(err).Debug("Failed Execution: failed running command")
		return errors.New("failed running command")
	}

	if task.Action.RequiresReceipt {
		if err := touchArtifact(ctx, task.Action.GeneratedResource, task.Action.Command); err != nil {
			logger.WithError(err).Debug("Failed Execution: failed uploading receipt")
			return errors.New("failed uploading receipt")
		}
	}

	logger.Debug("Successful Execution:")
	return nil
}

func touchArtifact(ctx context.Context, artifactName, action string) error {
	client, err := connection.NewClient(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to get client")
	}

	messageData, _ := proto.Marshal(&rpc.Receipt{Action: action})
	return core.SetArtifact(ctx, client, &rpc.Artifact{
		Name:     artifactName,
		Contents: messageData,
	})
}
