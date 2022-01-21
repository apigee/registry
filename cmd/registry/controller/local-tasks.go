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
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

// Implement io.Writer interface https://pkg.go.dev/io#Writer
type logWriter struct {
	logger log.Logger
}

func (w logWriter) Write(p []byte) (n int, err error) {
	w.logger.Debug(string(p))
	return len(p), nil
}

type ExecCommandTask struct {
	Action *Action
	TaskID string
	CreateCommand core.CommandGenerator
}

func (task *ExecCommandTask) String() string {
	return "Execute command: " + task.Action.Command
}

func (task *ExecCommandTask) Run(ctx context.Context) error {
	// The monitoring metrics/dashboards are built on top of the format of the log messages here.
	// Check the metric filters before making any changes to the format.
	// Location: registry/deployments/controller/dashboard/*
	logger := log.FromContext(ctx).WithFields(map[string]interface{}{
		"action": fmt.Sprintf("{%s}", task.Action.Command),
		"taskID": fmt.Sprintf("{%s}", task.TaskID),
	})

	err := task.ExcecuteCommand(ctx, logger)
	if err != nil {
		return err
	}

	logger.Debug("Successful Execution:")
	return nil
}

func (task *ExecCommandTask) ExcecuteCommand(ctx context.Context, logger log.Logger) error {

	if strings.HasPrefix(task.Action.Command, "registry resolve") {
		logger.Debug("Failed Execution: 'registry resolve' not allowed in action")
		return errors.New("'registry resolve' not allowed in action")
	}

	if task.CreateCommand == nil {
		task.CreateCommand = core.GetCommandGenerator()
	}

	fullCmd := strings.Fields(task.Action.Command)
	cmd := task.CreateCommand(fullCmd[0], fullCmd[1:]...)

	// Route the output appropriately for registry tool and third party commands

	// first party registry commands
	if strings.HasPrefix(task.Action.Command, "registry") {
		
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	} else { //third party commands
		// redirect the output of the subcommands to the logger
		cmdLogger := &logWriter{logger: logger}
		cmd.Stdout, cmd.Stderr = cmdLogger, cmdLogger
	}

	if err := cmd.Run(); err != nil {
		logger.WithError(err).Debug("Failed Execution: failed running command")
		return errors.New("failed running command")
	}

	if task.Action.RequiresReceipt {
		if err := task.touchArtifact(ctx, task.Action.GeneratedResource, task.Action.Command); err != nil {
			logger.WithError(err).Debug("Failed Execution: finished executing command, failed uploading receipt")
			return errors.New("executed command, failed uploading receipt")
		}
	}

	return nil
}

func (task *ExecCommandTask) touchArtifact(ctx context.Context, artifactName, action string) error {
	client, err := connection.NewClient(ctx)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
	}

	messageData, err := proto.Marshal(&rpc.Receipt{Action: action})
	if err != nil {
		return err
	}
	return core.SetArtifact(ctx, client, &rpc.Artifact{
		Name:     artifactName,
		MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.v1.controller.Receipt"),
		Contents: messageData,
	})
}
