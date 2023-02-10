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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"gopkg.in/yaml.v3"
)

func Apply(ctx context.Context, client connection.RegistryClient, path, project string, recursive bool, jobs int) error {
	patches := &patchGroup{}
	if path == "-" {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		if err := patches.parse(client, bytes, "", project); err != nil {
			return err
		}
		return patches.run(ctx, jobs)
	}

	err := filepath.WalkDir(path,
		func(fileName string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			} else if entry.IsDir() && fileName != path && !recursive {
				return filepath.SkipDir // Skip the directory and contents.
			} else if entry.IsDir() {
				return nil // Do nothing for the directory, but still walk its contents.
			} else if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".yml") {
				return nil // Skip everything that's not a YAML file.
			}
			bytes, err := os.ReadFile(fileName)
			if err != nil {
				return err
			}
			return patches.parse(client, bytes, fileName, project)
		})
	if err != nil {
		return err
	}
	return patches.run(ctx, jobs)
}

type patchGroup struct {
	apiTasks        []tasks.Task
	versionTasks    []tasks.Task
	specTasks       []tasks.Task
	deploymentTasks []tasks.Task
	artifactTasks   []tasks.Task
}

func (p *patchGroup) add(task *applyBytesTask) {
	switch task.kind {
	case "API":
		p.apiTasks = append(p.apiTasks, task)
	case "Version":
		p.versionTasks = append(p.versionTasks, task)
	case "Spec":
		p.specTasks = append(p.specTasks, task)
	case "Deployment":
		p.deploymentTasks = append(p.deploymentTasks, task)
	default: // for everything else, try an artifact type
		p.artifactTasks = append(p.artifactTasks, task)
	}
}

func (p *patchGroup) parse(client connection.RegistryClient, bytes []byte, fileName, project string) error {
	header, items, err := readHeaderWithItems(bytes)
	if err != nil {
		return err
	} else if header.ApiVersion != RegistryV1 {
		return nil
	}

	if items.Kind != yaml.SequenceNode {
		p.add(&applyBytesTask{
			client:  client,
			path:    fileName,
			project: project,
			parent:  header.Metadata.Parent,
			name:    header.Metadata.Name,
			kind:    header.Kind,
			bytes:   bytes,
		})
		return nil
	}

	for _, n := range items.Content {
		itemBytes, err := yaml.Marshal(n)
		if err != nil {
			return err
		}
		itemHeader, err := readHeader(itemBytes)
		if err != nil {
			return err
		}
		if itemHeader.ApiVersion != RegistryV1 {
			continue
		}
		p.add(&applyBytesTask{
			client:  client,
			path:    fileName,
			project: project,
			parent:  itemHeader.Metadata.Parent,
			name:    itemHeader.Metadata.Name,
			kind:    itemHeader.Kind,
			bytes:   itemBytes,
		})
	}

	return nil
}

func (p *patchGroup) run(ctx context.Context, jobs int) error {
	// Apply each resource type independently in order of ownership (parents first).
	for _, taskLists := range [][]tasks.Task{
		p.apiTasks,
		p.versionTasks,
		p.specTasks,
		p.deploymentTasks,
		p.artifactTasks,
	} {
		taskQueue, wait := tasks.WorkerPool(ctx, jobs)
		for _, task := range taskLists {
			taskQueue <- task
		}
		wait()
	}
	return nil
}

type applyBytesTask struct {
	client  connection.RegistryClient
	path    string
	project string
	parent  string
	name    string
	kind    string
	bytes   []byte
}

func (task *applyBytesTask) String() string {
	return fmt.Sprintf("apply %s (from path: %s)", task.resource(), task.path)
}

func (task *applyBytesTask) resource() string {
	var collection string
	switch task.kind {
	case "API":
		collection = "apis/"
	case "Version":
		collection = "versions/"
	case "Spec":
		collection = "specs/"
	case "Deployment":
		collection = "deployments/"
	default:
		collection = "artifacts/"
	}

	if task.parent == "" {
		return collection + task.name
	}
	return task.parent + "/" + collection + task.name
}

func (task *applyBytesTask) Run(ctx context.Context) error {
	header, err := readHeader(task.bytes)
	if err != nil {
		return err
	} else if header.ApiVersion != RegistryV1 {
		return nil
	}

	switch header.Kind {
	case "Version", "Spec", "Deployment":
		if header.Metadata.Parent == "" {
			return errors.New("metadata.parent should be set to the project-local parent path")
		}
	}

	log.FromContext(ctx).Infof("Applying %s", task.resource())
	switch header.Kind {
	case "API":
		return applyApiPatchBytes(ctx, task.client, task.bytes, task.project)
	case "Version":
		return applyApiVersionPatchBytes(ctx, task.client, task.bytes, task.project)
	case "Spec":
		return applyApiSpecPatchBytes(ctx, task.client, task.bytes, task.project, task.path)
	case "Deployment":
		return applyApiDeploymentPatchBytes(ctx, task.client, task.bytes, task.project)
	default: // for everything else, try an artifact type
		return applyArtifactPatchBytes(ctx, task.client, task.bytes, task.project)
	}
}
