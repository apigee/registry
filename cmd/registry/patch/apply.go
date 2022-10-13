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

func Apply(ctx context.Context, client connection.RegistryClient, path, parent string, recursive bool, taskQueue chan<- core.Task) error {
	return filepath.WalkDir(path,
		func(p string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			} else if entry.IsDir() && p != path && !recursive {
				return filepath.SkipDir // Skip the directory and contents.
			} else if entry.IsDir() {
				return nil // Do nothing for the directory, but still walk its contents.
			}
			return applyFile(ctx, client, p, parent, taskQueue)
		})
}

func applyFile(ctx context.Context, client connection.RegistryClient, fileName, parent string, taskQueue chan<- core.Task) error {
	if !strings.HasSuffix(fileName, ".yaml") {
		return nil
	}
	// Create an upload job for each API.
	taskQueue <- &applyFileTask{
		client: client,
		path:   fileName,
		parent: parent,
	}
	return nil
}

type applyFileTask struct {
	client connection.RegistryClient
	path   string
	parent string
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
	header, err := readHeader(bytes)
	if err != nil {
		return err
	}
	switch header.Kind {
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
