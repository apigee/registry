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

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"gopkg.in/yaml.v2"
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
		log.FromContext(ctx).WithError(err).Fatal("Failed to read file")
	}

	// get the id and kind of artifact from the YAML elements common to all artifacts
	var header Header
	err = yaml.Unmarshal(bytes, &header)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
	}
	if header.APIVersion != RegistryV1 {
		log.FromContext(ctx).Fatalf("Unsupported API version: %s", header.APIVersion)
	}
	switch header.Kind {
	case "API":
		var api API
		err = yaml.Unmarshal(bytes, &api)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
		}
		err = applyApiPatch(ctx, client, &api, parent)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
		}
	case "Lifecycle":
		var lifecycle Lifecycle
		err = yaml.Unmarshal(bytes, &lifecycle)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
		}
		err = applyArtifactPatch(ctx, client, &lifecycle, parent)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
		}
	case "Manifest":
		var manifest Manifest
		err = yaml.Unmarshal(bytes, &manifest)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
		}
		err = applyArtifactPatch(ctx, client, &manifest, parent)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
		}
	case "TaxonomyList":
		var taxonomyList TaxonomyList
		err = yaml.Unmarshal(bytes, &taxonomyList)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
		}
		err = applyArtifactPatch(ctx, client, &taxonomyList, parent)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
		}
	default:
		log.FromContext(ctx).Fatalf("Unsupported kind: %s", header.Kind)
	}
	return nil
}
