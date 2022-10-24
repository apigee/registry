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
)

type patchGroup struct {
	apiTasks        []*applyFileTask
	versionTasks    []*applyFileTask
	specTasks       []*applyFileTask
	deploymentTasks []*applyFileTask
	artifactTasks   []*applyFileTask
}

func newPatchGroup() *patchGroup {
	return &patchGroup{
		apiTasks:        make([]*applyFileTask, 0),
		versionTasks:    make([]*applyFileTask, 0),
		specTasks:       make([]*applyFileTask, 0),
		deploymentTasks: make([]*applyFileTask, 0),
		artifactTasks:   make([]*applyFileTask, 0),
	}
}

func (p *patchGroup) add(task *applyFileTask) error {
	bytes, err := os.ReadFile(task.path)
	if err != nil {
		return err
	}
	header, err := readHeader(bytes)
	if err != nil {
		return err
	}
	task.kind = header.Kind
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
	return nil
}

func (p *patchGroup) run(ctx context.Context, jobs int) error {
	log.FromContext(ctx).Infof("Applying Patch Group %+v", p)

	if len(p.apiTasks) > 0 {
		taskQueue, wait := core.WorkerPool(ctx, 64)
		for _, task := range p.apiTasks {
			taskQueue <- task
			wait()
		}
	}
	if len(p.versionTasks) > 0 {
		taskQueue, wait := core.WorkerPool(ctx, 64)
		for _, task := range p.versionTasks {
			taskQueue <- task
			wait()
		}
	}
	if len(p.specTasks) > 0 {
		taskQueue, wait := core.WorkerPool(ctx, 64)
		for _, task := range p.specTasks {
			taskQueue <- task
		}
		wait()
	}
	if len(p.deploymentTasks) > 0 {
		taskQueue, wait := core.WorkerPool(ctx, 64)
		for _, task := range p.deploymentTasks {
			taskQueue <- task
		}
		wait()
	}
	if len(p.artifactTasks) > 0 {
		taskQueue, wait := core.WorkerPool(ctx, 64)
		for _, task := range p.artifactTasks {
			taskQueue <- task
		}
		wait()
	}
	return nil
}

func Apply(ctx context.Context, client connection.RegistryClient, path, parent string, recursive bool, jobs int) error {
	patches := newPatchGroup()
	err := filepath.WalkDir(path,
		func(p string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			} else if entry.IsDir() && p != path && !recursive {
				return filepath.SkipDir // Skip the directory and contents.
			} else if entry.IsDir() {
				return nil // Do nothing for the directory, but still walk its contents.
			}
			return applyFile(ctx, client, p, parent, patches)
		})
	if err != nil {
		return err
	}
	return patches.run(ctx, jobs)
}

func applyFile(ctx context.Context, client connection.RegistryClient, fileName, parent string, patches *patchGroup) error {
	if !strings.HasSuffix(fileName, ".yaml") {
		return nil
	}
	return patches.add(&applyFileTask{
		client: client,
		path:   fileName,
		parent: parent,
	})
}

type applyFileTask struct {
	client connection.RegistryClient
	path   string
	parent string
	kind   string
}

func (task *applyFileTask) String() string {
	return "apply file " + task.path
}

func (task *applyFileTask) Run(ctx context.Context) error {
	log.FromContext(ctx).Infof("Applying %s", task.path)
	bytes, err := os.ReadFile(task.path)
	if err != nil {
		return err
	}
	switch task.kind {
	case "API":
		return applyApiPatchBytes(ctx, task.client, bytes, task.parent)
	case "Version":
		return applyApiVersionPatchBytes(ctx, task.client, bytes, task.parent)
	case "Spec":
		return applyApiSpecPatchBytes(ctx, task.client, bytes, task.parent)
	case "Deployment":
		return applyApiDeploymentPatchBytes(ctx, task.client, bytes, task.parent)
	default: // for everything else, try an artifact type
		return applyArtifactPatchBytes(ctx, task.client, bytes, task.parent)
	}
}
