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
	"io/ioutil"
	"path/filepath"

	"github.com/apigee/registry/connection"
)

func Apply(
	ctx context.Context,
	client connection.Client,
	fileName string,
	recursive bool,
	parent string) error {
	return filepath.WalkDir(fileName,
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if !recursive {
					return filepath.SkipDir
				}
				return nil
			}
			return applyFile(ctx, client, path, parent)
		})
}

func applyFile(
	ctx context.Context,
	client connection.Client,
	fileName string,
	parent string) error {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	// get the id and kind of artifact from the YAML elements common to all artifacts
	header, err := readHeader(bytes)
	if err != nil {
		return err
	}
	switch header.Kind {
	case "API":
		return applyApiPatch(ctx, client, bytes, parent)
	case "Lifecycle":
		return applyLifecycleArtifactPatch(ctx, client, bytes, parent)
	case "Manifest":
		return applyManifestArtifactPatch(ctx, client, bytes, parent)
	case "TaxonomyList":
		return applyTaxonomyListArtifactPatch(ctx, client, bytes, parent)
	default:
		return fmt.Errorf("Unsupported kind: %s", header.Kind)
	}
}
