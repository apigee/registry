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
	"os"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"gopkg.in/yaml.v2"
)

// ExportProject writes a project as a YAML file.
func ExportProject(ctx context.Context, client *gapic.RegistryClient, adminClient *gapic.AdminClient, message *rpc.Project) {
	projectName, err := names.ParseProject(message.Name)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to parse project name")
	}
	apisdir := "apis"
	err = os.MkdirAll(apisdir, 0777)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to create output directory")
	}
	core.ListAPIs(ctx, client, projectName.Api(""), "", func(message *rpc.Api) {
		api, err := buildAPI(ctx, client, message)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to export api as YAML")
			return
		}
		b, err := yaml.Marshal(api)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to marshal api as YAML")
			return
		}
		filename := fmt.Sprintf("%s/%s.yaml", apisdir, api.Header.Metadata.Name)
		err = os.WriteFile(filename, b, 0644)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to write output YAML")
		}
	})
}

// ExportAPI writes an API as a YAML file.
func ExportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) {
	api, err := buildAPI(ctx, client, message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export api")
	}
	b, err := yaml.Marshal(api)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportArtifact writes an artifact as a YAML file.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) {
	artifact, err := buildArtifact(message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
	}
	b, err := yaml.Marshal(artifact)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}
