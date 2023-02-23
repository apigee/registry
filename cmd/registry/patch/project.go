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
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

// NewProject gets a serialized representation of a project.
func NewProject(ctx context.Context, client *gapic.RegistryClient, message *rpc.Project) (*encoding.Project, error) {
	projectName, err := names.ParseProject(message.Name)
	if err != nil {
		return nil, err
	}
	return &encoding.Project{
		Header: encoding.Header{
			ApiVersion: encoding.RegistryV1,
			Kind:       "Project",
			Metadata: encoding.Metadata{
				Name: projectName.ProjectID,
			},
		},
		Data: encoding.ProjectData{
			DisplayName: message.DisplayName,
			Description: message.Description,
		},
	}, err
}

func applyProjectPatchBytes(ctx context.Context, client connection.AdminClient, bytes []byte) error {
	var project encoding.Project
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
	return err
}

// ExportProject writes a project into a directory of YAML files.
func ExportProject(ctx context.Context, client *gapic.RegistryClient, projectName names.Project, root string, taskQueue chan<- tasks.Task) error {
	root = filepath.Join(root, projectName.ProjectID)
	if err := visitor.ListAPIs(ctx, client, projectName.Api(""), "", func(ctx context.Context, message *rpc.Api) error {
		taskQueue <- &exportAPITask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListArtifacts(ctx, client, projectName.Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListVersions(ctx, client, projectName.Api("-").Version("-"), "", func(ctx context.Context, message *rpc.ApiVersion) error {
		taskQueue <- &exportVersionTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListSpecs(ctx, client, projectName.Api("-").Version("-").Spec("-"), "", false, func(ctx context.Context, message *rpc.ApiSpec) error {
		taskQueue <- &exportSpecTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListDeployments(ctx, client, projectName.Api("-").Deployment("-"), "", func(ctx context.Context, message *rpc.ApiDeployment) error {
		taskQueue <- &exportDeploymentTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListArtifacts(ctx, client, projectName.Api("-").Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListArtifacts(ctx, client, projectName.Api("-").Version("-").Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListArtifacts(ctx, client, projectName.Api("-").Version("-").Spec("-").Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	return visitor.ListArtifacts(ctx, client, projectName.Api("-").Deployment("-").Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
}

// ExportAPI writes an API into a directory of YAML files.
func ExportAPI(ctx context.Context, client *gapic.RegistryClient, apiName names.Api, recursive bool, root string, taskQueue chan<- tasks.Task) error {
	root = filepath.Join(root, apiName.ProjectID)
	if err := visitor.ListAPIs(ctx, client, apiName, "", func(ctx context.Context, message *rpc.Api) error {
		taskQueue <- &exportAPITask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if !recursive {
		return nil
	}
	if err := visitor.ListVersions(ctx, client, apiName.Version("-"), "", func(ctx context.Context, message *rpc.ApiVersion) error {
		taskQueue <- &exportVersionTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListSpecs(ctx, client, apiName.Version("-").Spec("-"), "", false, func(ctx context.Context, message *rpc.ApiSpec) error {
		taskQueue <- &exportSpecTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListDeployments(ctx, client, apiName.Deployment("-"), "", func(ctx context.Context, message *rpc.ApiDeployment) error {
		taskQueue <- &exportDeploymentTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListArtifacts(ctx, client, apiName.Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListArtifacts(ctx, client, apiName.Version("-").Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListArtifacts(ctx, client, apiName.Version("-").Spec("-").Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	return visitor.ListArtifacts(ctx, client, apiName.Deployment("-").Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
}

// ExportAPIVersion writes an API version into a directory of YAML files.
func ExportAPIVersion(ctx context.Context, client *gapic.RegistryClient, versionName names.Version, recursive bool, root string, taskQueue chan<- tasks.Task) error {
	root = filepath.Join(root, versionName.ProjectID)
	if err := visitor.ListVersions(ctx, client, versionName, "", func(ctx context.Context, message *rpc.ApiVersion) error {
		taskQueue <- &exportVersionTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if !recursive {
		return nil
	}
	if err := visitor.ListSpecs(ctx, client, versionName.Spec("-"), "", false, func(ctx context.Context, message *rpc.ApiSpec) error {
		taskQueue <- &exportSpecTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if err := visitor.ListArtifacts(ctx, client, versionName.Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	return visitor.ListArtifacts(ctx, client, versionName.Spec("-").Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
}

// ExportAPISpec writes an API spec into a directory of YAML files.
func ExportAPISpec(ctx context.Context, client *gapic.RegistryClient, specName names.Spec, recursive bool, root string, taskQueue chan<- tasks.Task) error {
	root = filepath.Join(root, specName.ProjectID)
	if err := visitor.ListSpecs(ctx, client, specName, "", false, func(ctx context.Context, message *rpc.ApiSpec) error {
		taskQueue <- &exportSpecTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if !recursive {
		return nil
	}
	return visitor.ListArtifacts(ctx, client, specName.Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
}

// ExportAPIDeployment writes an API deployment into a directory of YAML files.
func ExportAPIDeployment(ctx context.Context, client *gapic.RegistryClient, deploymentName names.Deployment, recursive bool, root string, taskQueue chan<- tasks.Task) error {
	root = filepath.Join(root, deploymentName.ProjectID)
	if err := visitor.ListDeployments(ctx, client, deploymentName, "", func(ctx context.Context, message *rpc.ApiDeployment) error {
		taskQueue <- &exportDeploymentTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	}); err != nil {
		return err
	}
	if !recursive {
		return nil
	}
	return visitor.ListArtifacts(ctx, client, deploymentName.Artifact(""), "", false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
		}
		return nil
	})
}

// ExportArtifact writes an artifact into a directory of YAML files.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, artifactName names.Artifact, root string, taskQueue chan<- tasks.Task) error {
	root = filepath.Join(root, artifactName.ProjectID())
	return visitor.GetArtifact(ctx, client, artifactName, false, func(ctx context.Context, message *rpc.Artifact) error {
		taskQueue <- &exportArtifactTask{
			client:  client,
			message: message,
			dir:     root,
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
	api, err := NewApi(ctx, task.client, task.message, false)
	if err != nil {
		return err
	}
	bytes, err := encoding.EncodeYAML(api)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	filename := fmt.Sprintf("%s/apis/%s/info.yaml", task.dir, api.Header.Metadata.Name)
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
	artifact, err := NewArtifact(ctx, task.client, task.message)
	if err != nil {
		log.FromContext(ctx).Warnf("Skipped %s: %s", task.message.Name, err)
		return nil
	}
	if artifact.Header.Kind == "Artifact" { // "Artifact" is the generic artifact type
		log.FromContext(ctx).Warnf("Skipped %s", task.message.Name)
		return nil
	}
	bytes, err := encoding.EncodeYAML(artifact)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	var filename string
	if artifact.Header.Metadata.Parent == "" {
		filename = fmt.Sprintf("%s/artifacts/%s.yaml", task.dir, artifact.Header.Metadata.Name)
	} else {
		filename = fmt.Sprintf("%s/%s/artifacts/%s.yaml", task.dir, artifact.Header.Metadata.Parent, artifact.Header.Metadata.Name)
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
}

func (task *exportVersionTask) String() string {
	return "export " + task.message.Name
}

func (task *exportVersionTask) Run(ctx context.Context) error {
	version, err := NewApiVersion(ctx, task.client, task.message, false)
	if err != nil {
		return err
	}
	bytes, err := encoding.EncodeYAML(version)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	filename := fmt.Sprintf("%s/%s/versions/%s/info.yaml", task.dir, version.Header.Metadata.Parent, version.Header.Metadata.Name)
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
}

func (task *exportSpecTask) String() string {
	return "export " + task.message.Name
}

func (task *exportSpecTask) Run(ctx context.Context) error {
	spec, err := NewApiSpec(ctx, task.client, task.message, false)
	if err != nil {
		return err
	}
	bytes, err := encoding.EncodeYAML(spec)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	filename := fmt.Sprintf("%s/%s/specs/%s/info.yaml", task.dir, spec.Header.Metadata.Parent, spec.Header.Metadata.Name)
	parentDir := filepath.Dir(filename)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		return err
	}
	if err = os.WriteFile(filename, bytes, 0644); err != nil {
		return err
	}
	if task.message.Filename == "" {
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
		data, _ = compress.GUnzippedBytes(data)
	}
	return os.WriteFile(filepath.Join(parentDir, task.message.Filename), data, 0644)
}

type exportDeploymentTask struct {
	client  connection.RegistryClient
	message *rpc.ApiDeployment
	dir     string
}

func (task *exportDeploymentTask) String() string {
	return "export " + task.message.Name
}

func (task *exportDeploymentTask) Run(ctx context.Context) error {
	deployment, err := NewApiDeployment(ctx, task.client, task.message, false)
	if err != nil {
		return err
	}
	bytes, err := encoding.EncodeYAML(deployment)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Infof("Exported %s", task.message.Name)
	filename := fmt.Sprintf("%s/%s/deployments/%s/info.yaml", task.dir, deployment.Header.Metadata.Parent, deployment.Header.Metadata.Name)
	parentDir := filepath.Dir(filename)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, 0644)
}
