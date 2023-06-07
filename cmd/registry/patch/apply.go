// Copyright 2022 Google LLC.
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
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/log"
	"gopkg.in/yaml.v3"
)

func Apply(ctx context.Context, client connection.RegistryClient, adminClient connection.AdminClient, in io.Reader, project string, recursive bool, jobs int, paths ...string) error {
	patches := &patchGroup{}
	if paths[0] == "-" {
		bytes, err := io.ReadAll(in)
		if err != nil {
			return err
		}
		if err := patches.parse(client, adminClient, bytes, "", project); err != nil {
			return err
		}
		return patches.run(ctx, jobs)
	}

	for _, p := range paths {
		err := filepath.WalkDir(p,
			func(fileName string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				} else if entry.IsDir() && fileName != p && !recursive {
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
				err = patches.parse(client, adminClient, bytes, fileName, project)
				if err != nil {
					err = fmt.Errorf("parsing %s: %w", fileName, err)
				}
				return err
			})
		if err != nil {
			return err
		}
	}
	return patches.run(ctx, jobs)
}

type patchGroup struct {
	filesRead       int
	filesApplied    int
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

func (p *patchGroup) parse(client connection.RegistryClient, adminClient connection.AdminClient, bytes []byte, fileName, project string) error {
	p.filesRead++
	header, items, err := readHeaderWithItems(bytes)
	if err != nil {
		return err
	} else if header.ApiVersion != encoding.RegistryV1 {
		return nil
	}
	p.filesApplied++
	if items.Kind != yaml.SequenceNode {
		p.add(&applyBytesTask{
			client:      client,
			adminClient: adminClient,
			path:        fileName,
			project:     project,
			parent:      header.Metadata.Parent,
			name:        header.Metadata.Name,
			kind:        header.Kind,
			bytes:       bytes,
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
		if itemHeader.ApiVersion != encoding.RegistryV1 {
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
		taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
		for _, task := range taskLists {
			taskQueue <- task
		}
		wait()
	}
	if p.filesRead == 0 {
		return fmt.Errorf("no YAML files found")
	}
	if p.filesApplied == 0 {
		return fmt.Errorf("no YAML files applied (%d found, none with 'apiVersion: %s')", p.filesRead, encoding.RegistryV1)
	}
	log.FromContext(ctx).Infof("%d YAML file(s) applied (%d found, %d with 'apiVersion: %s')", p.filesApplied, p.filesRead, p.filesApplied, encoding.RegistryV1)
	return nil
}

type applyBytesTask struct {
	client      connection.RegistryClient
	adminClient connection.AdminClient
	path        string
	project     string
	parent      string
	name        string
	kind        string
	bytes       []byte
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
	} else if header.ApiVersion != encoding.RegistryV1 {
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
	case "Project":
		return applyProjectPatchBytes(ctx, task.adminClient, task.bytes)
	case "API":
		return applyApiPatchBytes(ctx, task.client, task.bytes, task.project, task.path)
	case "Version":
		return applyApiVersionPatchBytes(ctx, task.client, task.bytes, task.project, task.path)
	case "Spec":
		return applyApiSpecPatchBytes(ctx, task.client, task.bytes, task.project, task.path)
	case "Deployment":
		return applyApiDeploymentPatchBytes(ctx, task.client, task.bytes, task.project, task.path)
	default: // for everything else, try an artifact type
		return applyArtifactPatchBytes(ctx, task.client, task.bytes, task.project, task.path)
	}
}
