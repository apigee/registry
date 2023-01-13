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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"gopkg.in/yaml.v3"
)

func Apply(ctx context.Context, client connection.RegistryClient, path, parent string, recursive bool, jobs int) error {
	patches := &patchGroup{}
	err := filepath.WalkDir(path,
		func(fileName string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			} else if entry.IsDir() && fileName != path && !recursive {
				return filepath.SkipDir // Skip the directory and contents.
			} else if entry.IsDir() {
				return nil // Do nothing for the directory, but still walk its contents.
			} else if !strings.HasSuffix(fileName, ".yaml") {
				return nil // Skip everything that's not a YAML file.
			}
			bytes, err := os.ReadFile(fileName)
			if err != nil {
				return err
			}
			header, items, err := readHeaderWithItems(bytes)
			if err != nil {
				return err
			}
			if header.ApiVersion != RegistryV1 {
				return nil
			}
			if items.Kind == yaml.SequenceNode {
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
						return nil
					}
					patches.add(&applyBytesTask{
						client: client,
						path:   fileName,
						parent: parent,
						kind:   itemHeader.Kind,
						bytes:  itemBytes,
					})
				}
				return nil
			}
			patches.add(&applyBytesTask{
				client: client,
				path:   fileName,
				parent: parent,
				kind:   header.Kind,
				bytes:  bytes,
			})
			return nil
		})
	if err != nil {
		return err
	}
	return patches.run(ctx, jobs)
}

type patchGroup struct {
	apiTasks        []core.Task
	versionTasks    []core.Task
	specTasks       []core.Task
	deploymentTasks []core.Task
	artifactTasks   []core.Task
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

func (p *patchGroup) run(ctx context.Context, jobs int) error {
	// Apply each resource type independently in order of ownership (parents first).
	for _, tasks := range [][]core.Task{
		p.apiTasks,
		p.versionTasks,
		p.specTasks,
		p.deploymentTasks,
		p.artifactTasks,
	} {
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		for _, task := range tasks {
			taskQueue <- task
		}
		wait()
	}
	return nil
}

type applyBytesTask struct {
	client connection.RegistryClient
	path   string
	parent string
	kind   string
	bytes  []byte
}

func (task *applyBytesTask) String() string {
	return "apply " + task.path
}

func (task *applyBytesTask) Run(ctx context.Context) error {
	header, err := readHeader(task.bytes)
	if err != nil {
		return err
	}
	if header.ApiVersion != RegistryV1 {
		return nil
	}
	name := header.Metadata.Name
	if header.Metadata.Parent != "" {
		name = header.Metadata.Parent + "/" + name
	}
	log.FromContext(ctx).Infof("Applying %s %s", task.path, name)
	switch header.Kind {
	case "API":
		return applyApiPatchBytes(ctx, task.client, task.bytes, task.parent)
	case "Version":
		return applyApiVersionPatchBytes(ctx, task.client, task.bytes, task.parent)
	case "Spec":
		return applyApiSpecPatchBytes(ctx, task.client, task.bytes, task.parent, task.path)
	case "Deployment":
		return applyApiDeploymentPatchBytes(ctx, task.client, task.bytes, task.parent)
	default: // for everything else, try an artifact type
		return applyArtifactPatchBytes(ctx, task.client, task.bytes, task.parent)
	}
}
