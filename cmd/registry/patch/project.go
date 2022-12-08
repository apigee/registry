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

package patch

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

// ExportProject writes a project into a directory of YAML files.
func ExportProject(ctx context.Context, client *gapic.RegistryClient, projectName names.Project, taskQueue chan<- core.Task) error {
	apisDir := fmt.Sprintf("%s/apis", projectName.ProjectID)
	if err := os.MkdirAll(apisDir, 0777); err != nil {
		return err
	}
	err := core.ListAPIs(ctx, client, projectName.Api(""), "", func(message *rpc.Api) error {
		taskQueue <- &exportAPITask{
			client:  client,
			message: message,
			dir:     apisDir,
		}
		return nil
	})
	if err != nil {
		return err
	}

	artifactsDir := fmt.Sprintf("%s/artifacts", projectName.ProjectID)
	if err := os.MkdirAll(artifactsDir, 0777); err != nil {
		return err
	}

	return core.ListArtifacts(ctx, client, projectName.Artifact(""), "", false, func(message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     artifactsDir,
		}
		return nil
	})
}

type exportAPITask struct {
	client  connection.RegistryClient
	message *rpc.Api
	dir     string
}

func (task *exportAPITask) String() string {
	return "export " + task.message.Name
}

func (task *exportAPITask) Run(ctx context.Context) error {
	bytes, header, err := ExportAPI(ctx, task.client, task.message, true)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	filename := fmt.Sprintf("%s/%s.yaml", task.dir, header.Metadata.Name)
	return os.WriteFile(filename, bytes, 0644)
}

type exportArtifactTask struct {
	client  connection.RegistryClient
	message *rpc.Artifact
	dir     string
}

func (task *exportArtifactTask) String() string {
	return "export " + task.message.Name
}

func (task *exportArtifactTask) Run(ctx context.Context) error {
	bytes, header, err := ExportArtifact(ctx, task.client, task.message)
	if err != nil {
		log.FromContext(ctx).Warnf("Skipped %s: %s", task.message.Name, err)
		return nil
	}
	if header.Kind == "Artifact" { // "Artifact" is the generic artifact type
		log.FromContext(ctx).Warnf("Skipped %s", task.message.Name)
		return nil
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	filename := fmt.Sprintf("%s/%s.yaml", task.dir, header.Metadata.Name)
	return os.WriteFile(filename, bytes, fs.ModePerm)
}
