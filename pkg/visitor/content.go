// Copyright 2020 Google LLC.
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
	"path"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func SetArtifact(ctx context.Context,
	client *gapic.RegistryClient,
	artifact *rpc.Artifact) error {
	request := &rpc.CreateArtifactRequest{}
	request.Artifact = artifact
	request.ArtifactId = path.Base(artifact.GetName())
	request.Parent = path.Dir(path.Dir(artifact.GetName()))
	// First try setting a new artifact value.
	_, err := client.CreateArtifact(ctx, request)
	if err == nil {
		return nil
	}
	// If that failed because the artifact already exists, replace it.
	code := status.Code(err)
	if code == codes.AlreadyExists {
		request := &rpc.ReplaceArtifactRequest{}
		request.Artifact = artifact
		_, err = client.ReplaceArtifact(ctx, request)
	}
	return err
}
