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
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/models"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

func newProject(ctx context.Context, client *gapic.RegistryClient, message *rpc.Project) (*models.Project, error) {
	projectName, err := names.ParseProject(message.Name)
	if err != nil {
		return nil, err
	}
	return &models.Project{
		Header: models.Header{
			ApiVersion: RegistryV1,
			Kind:       "Project",
			Metadata: models.Metadata{
				Name: projectName.ProjectID,
			},
		},
		Data: models.ProjectData{
			DisplayName: message.DisplayName,
			Description: message.Description,
		},
	}, err
}

// PatchForProject gets a serialized representation of a project.
func PatchForProject(ctx context.Context, client *gapic.RegistryClient, message *rpc.Project) ([]byte, *models.Header, error) {
	project, err := newProject(ctx, client, message)
	if err != nil {
		return nil, nil, err
	}
	var b bytes.Buffer
	err = yamlEncoder(&b).Encode(project)
	if err != nil {
		return nil, nil, err
	}
	return b.Bytes(), &project.Header, nil
}

func applyProjectPatchBytes(ctx context.Context, client connection.AdminClient, bytes []byte) error {
	var project models.Project
	err := yaml.Unmarshal(bytes, &project)
	if err != nil {
		return err
	}
	req := &rpc.UpdateProjectRequest{
		Project: &rpc.Project{
			Name:        "projects/" + project.Metadata.Name,
			DisplayName: project.Data.DisplayName,
			Description: project.Data.Description,
		},
		AllowMissing: true,
	}
	_, err = client.UpdateProject(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

// ExportProject writes a project into a directory of YAML files.
func ExportProject(ctx context.Context, client *gapic.RegistryClient, projectName names.Project, root string, taskQueue chan<- core.Task, nested bool) error {
	if root != "" {
		root = root + "/" + projectName.ProjectID
	} else {
		root = projectName.ProjectID
	}
	err := core.ListAPIs(ctx, client, projectName.Api(""), "", func(message *rpc.Api) error {
		taskQueue <- &exportAPITask{
			client:  client,
			message: message,
			dir:     root,
			nested:  nested,
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = core.ListArtifacts(ctx, client, projectName.Artifact(""), "", false, func(message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
	if err != nil {
		return err
	}
	if nested {
		return nil
	}
	err = core.ListVersions(ctx, client, projectName.Api("-").Version("-"), "", func(message *rpc.ApiVersion) error {
		taskQueue <- &exportVersionTask{
			client:  client,
			message: message,
			dir:     root,
			nested:  nested,
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = core.ListSpecs(ctx, client, projectName.Api("-").Version("-").Spec("-"), "", false, func(message *rpc.ApiSpec) error {
		taskQueue <- &exportSpecTask{
			client:  client,
			message: message,
			dir:     root,
			nested:  nested,
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = core.ListDeployments(ctx, client, projectName.Api("-").Deployment("-"), "", func(message *rpc.ApiDeployment) error {
		taskQueue <- &exportDeploymentTask{
			client:  client,
			message: message,
			dir:     root,
			nested:  nested,
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = core.ListArtifacts(ctx, client, projectName.Api("-").Artifact(""), "", false, func(message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = core.ListArtifacts(ctx, client, projectName.Api("-").Version("-").Artifact(""), "", false, func(message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = core.ListArtifacts(ctx, client, projectName.Api("-").Version("-").Spec("-").Artifact(""), "", false, func(message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = core.ListArtifacts(ctx, client, projectName.Api("-").Deployment("-").Artifact(""), "", false, func(message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

type exportAPITask struct {
	client  connection.RegistryClient
	message *rpc.Api
	dir     string
	nested  bool
}

func (task *exportAPITask) String() string {
	return "export " + task.message.Name
}

func (task *exportAPITask) Run(ctx context.Context) error {
	bytes, header, err := PatchForApi(ctx, task.client, task.message, task.nested)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	var filename string
	if task.nested {
		filename = fmt.Sprintf("%s/apis/%s.yaml", task.dir, header.Metadata.Name)
	} else {
		filename = fmt.Sprintf("%s/apis/%s/info.yaml", task.dir, header.Metadata.Name)
	}
	parentDir := filepath.Dir(filename)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		return err
	}
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
	bytes, header, err := PatchForArtifact(ctx, task.client, task.message)
	if err != nil {
		log.FromContext(ctx).Warnf("Skipped %s: %s", task.message.Name, err)
		return nil
	}
	if header.Kind == "Artifact" { // "Artifact" is the generic artifact type
		log.FromContext(ctx).Warnf("Skipped %s", task.message.Name)
		return nil
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	var filename string
	if header.Metadata.Parent == "" {
		filename = fmt.Sprintf("%s/artifacts/%s.yaml", task.dir, header.Metadata.Name)
	} else {
		filename = fmt.Sprintf("%s/%s/artifacts/%s.yaml", task.dir, header.Metadata.Parent, header.Metadata.Name)
	}
	parentDir := filepath.Dir(filename)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, fs.ModePerm)
}

type exportVersionTask struct {
	client  connection.RegistryClient
	message *rpc.ApiVersion
	dir     string
	nested  bool
}

func (task *exportVersionTask) String() string {
	return "export " + task.message.Name
}

func (task *exportVersionTask) Run(ctx context.Context) error {
	bytes, header, err := PatchForApiVersion(ctx, task.client, task.message, task.nested)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	var filename string
	if task.nested {
		filename = ""
	} else {
		filename = fmt.Sprintf("%s/%s/versions/%s/info.yaml", task.dir, header.Metadata.Parent, header.Metadata.Name)
	}
	parentDir := filepath.Dir(filename)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, 0644)
}

type exportSpecTask struct {
	client  connection.RegistryClient
	message *rpc.ApiSpec
	dir     string
	nested  bool
}

func (task *exportSpecTask) String() string {
	return "export " + task.message.Name
}

func (task *exportSpecTask) Run(ctx context.Context) error {
	bytes, header, err := PatchForApiSpec(ctx, task.client, task.message, task.nested)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	var filename string
	if task.nested {
		filename = ""
	} else {
		filename = fmt.Sprintf("%s/%s/specs/%s/info.yaml", task.dir, header.Metadata.Parent, header.Metadata.Name)
	}
	parentDir := filepath.Dir(filename)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		return err
	}
	if err = os.WriteFile(filename, bytes, 0644); err != nil {
		return err
	}
	if task.nested {
		return nil
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "accept-encoding", "gzip")
	contents, err := task.client.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
		Name: task.message.GetName(),
	})
	if err != nil {
		return err
	}
	data := contents.GetData()
	if strings.Contains(contents.GetContentType(), "+gzip") {
		data, _ = core.GUnzippedBytes(data)
	}
	return os.WriteFile(filepath.Join(parentDir, task.message.Filename), data, 0644)
}

type exportDeploymentTask struct {
	client  connection.RegistryClient
	message *rpc.ApiDeployment
	dir     string
	nested  bool
}

func (task *exportDeploymentTask) String() string {
	return "export " + task.message.Name
}

func (task *exportDeploymentTask) Run(ctx context.Context) error {
	bytes, header, err := PatchForApiDeployment(ctx, task.client, task.message, task.nested)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	var filename string
	if task.nested {
		filename = ""
	} else {
		filename = fmt.Sprintf("%s/%s/deployments/%s/info.yaml", task.dir, header.Metadata.Parent, header.Metadata.Name)
	}
	parentDir := filepath.Dir(filename)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, 0644)
}
