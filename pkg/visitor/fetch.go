// Copyright 2023 Google LLC.
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

package visitor

import (
	"context"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/metadata"
)

func FetchSpecContents(ctx context.Context, client *gapic.RegistryClient, spec *rpc.ApiSpec) error {
	if spec.Contents != nil {
		return nil
	}
	request := &rpc.GetApiSpecContentsRequest{
		Name: spec.GetName(),
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "accept-encoding", "gzip")
	contents, err := client.GetApiSpecContents(ctx, request)
	if err != nil {
		return err
	}
	spec.MimeType = contents.GetContentType()
	spec.Contents = contents.GetData()
	if mime.IsGZipCompressed(spec.MimeType) {
		spec.MimeType = mime.GUnzippedType(spec.MimeType)
		spec.Contents, err = compress.GUnzippedBytes(spec.Contents)
	}
	return err
}

func FetchArtifactContents(ctx context.Context, client *gapic.RegistryClient, artifact *rpc.Artifact) error {
	if artifact.Contents != nil {
		return nil
	}
	contents, err := client.GetArtifactContents(ctx, &rpc.GetArtifactContentsRequest{
		Name: artifact.GetName(),
	})
	if err != nil {
		return err
	}
	artifact.Contents = contents.GetData()
	artifact.MimeType = contents.GetContentType()
	return nil
}
