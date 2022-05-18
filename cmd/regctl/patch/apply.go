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
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
)

func Apply(ctx context.Context, client connection.Client, path, parent string, recursive bool) error {
	return filepath.WalkDir(path,
		func(p string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			} else if entry.IsDir() && p != path && !recursive {
				return filepath.SkipDir // Skip the directory and contents.
			} else if entry.IsDir() {
				return nil // Do nothing for the directory, but still walk its contents.
			}
			return applyFile(ctx, client, p, parent)
		})
}

func applyFile(ctx context.Context, client connection.Client, fileName, parent string) error {
	if !strings.HasSuffix(fileName, ".yaml") {
		return nil
	}
	log.FromContext(ctx).Infof("Importing %s", fileName)
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	header, err := readHeader(bytes)
	if err != nil {
		return err
	}
	switch header.Kind {
	case "API":
		return applyApiPatch(ctx, client, bytes, parent)
	default: // for everything else, try an artifact type
		return applyArtifactPatchBytes(ctx, client, bytes, parent)
	}
}
