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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

// ExportProject writes a project into a directory of YAML files.
func ExportProject(ctx context.Context, client *gapic.RegistryClient, name string) error {
	projectName, err := names.ParseProject(name)
	if err != nil {
		return err
	}
	apisDir := fmt.Sprintf("%s/apis", projectName.ProjectID)
	err = os.MkdirAll(apisDir, 0777)
	if err != nil {
		return err
	}
	err = core.ListAPIs(ctx, client, projectName.Api(""), "", func(message *rpc.Api) {
		log.FromContext(ctx).Infof("Exporting %s", message.Name)
		bytes, header, err := ExportAPI(ctx, client, message)
		if err != nil {
			// TODO: this should return an error instead of logging (requires ListAPIs modification)
			log.FromContext(ctx).WithError(err).Fatal("Failed to export API")
		}
		filename := fmt.Sprintf("%s/%s.yaml", apisDir, header.Metadata.Name)
		err = os.WriteFile(filename, bytes, 0644)
		if err != nil {
			// TODO: this should return an error instead of logging (requires ListAPIs modification)
			log.FromContext(ctx).WithError(err).Fatal("Failed to write output YAML")
		}
	})
	if err != nil {
		return err
	}
	artifactsDir := fmt.Sprintf("%s/artifacts", projectName.ProjectID)
	err = os.MkdirAll(artifactsDir, 0777)
	if err != nil {
		return err
	}
	err = core.ListArtifacts(ctx, client, projectName.Artifact(""), "", false, func(message *rpc.Artifact) {
		log.FromContext(ctx).Infof("Exporting %s", message.Name)
		bytes, header, err := ExportArtifact(ctx, client, message)
		if err != nil {
			// TODO: this should return an error instead of logging (requires ListArtifacts modification)
			log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
		}
		filename := fmt.Sprintf("%s/%s.yaml", artifactsDir, header.Metadata.Name)
		err = os.WriteFile(filename, bytes, fs.ModePerm)
		if err != nil {
			// TODO: this should return an error instead of logging (requires ListArtifacts modification)
			log.FromContext(ctx).WithError(err).Fatal("Failed to write output YAML")
		}
	})
	return err
}
