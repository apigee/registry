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

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

type Artifact interface {
	GetHeader() *Header
	GetMessage() proto.Message
	GetMimeType() string
}

// ExportArtifact allows an artifact to be individually exported as a YAML file.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) ([]byte, *Header, error) {
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
	b, err := marshalYAML(artifact)
	if err != nil {
		return nil, nil, err
	}
	return b, artifact.GetHeader(), nil
}

func applyArtifactPatch(
	ctx context.Context,
	client connection.Client,
	content Artifact,
	parent string) error {
	bytes, err := proto.Marshal(content.GetMessage())
	if err != nil {
		return err
	}
	artifact := &rpc.Artifact{
		Name:     fmt.Sprintf("%s/artifacts/%s", parent, content.GetHeader().Metadata.Name),
		MimeType: content.GetMimeType(),
		Contents: bytes,
	}
	req := &rpc.CreateArtifactRequest{
		Parent:     parent,
		ArtifactId: content.GetHeader().Metadata.Name,
		Artifact:   artifact,
	}
	_, err = client.CreateArtifact(ctx, req)
	if err != nil {
		req := &rpc.ReplaceArtifactRequest{
			Artifact: artifact,
		}
		_, err = client.ReplaceArtifact(ctx, req)
	}
	return err
}
