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

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

type Artifact interface {
	GetHeader() *Header
	GetMessage() proto.Message
}

// ExportArtifact writes an artifact as a YAML file.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) {
	bytes, _, err := exportArtifact(ctx, client, message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
	} else {
		fmt.Println(string(bytes))
	}
}

func exportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) ([]byte, *Header, error) {
	if message.Contents == nil {
		req := &rpc.GetArtifactContentsRequest{
			Name: message.Name,
		}
		body, err := client.GetArtifactContents(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		message.Contents = body.Data
	}
	var artifact Artifact
	var err error
	switch message.GetMimeType() {
	case LifecycleMimeType:
		artifact, err = newLifecycle(message)
	case ManifestMimeType:
		artifact, err = newManifest(message)
	case TaxonomyListMimeType:
		artifact, err = newTaxonomyList(message)
	default:
		artifact, err = newUnknownArtifact(message)
	}
	if err != nil {
		return nil, nil, err
	}
	b, err := yaml.Marshal(artifact)
	if err != nil {
		return nil, nil, err
	}
	return b, artifact.GetHeader(), nil
}
