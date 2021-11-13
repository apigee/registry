// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

func ListResources(ctx context.Context, client connection.Client, pattern, filter string) ([]ResourceInstance, error) {
	var result []ResourceInstance
	var err2 error

	// First try to match collection names.
	if api, err := names.ParseApiCollection(pattern); err == nil {
		err2 = core.ListAPIs(ctx, client, api, filter, generateApiHandler(&result))
	} else if version, err := names.ParseVersionCollection(pattern); err == nil {
		err2 = core.ListVersions(ctx, client, version, filter, generateVersionHandler(&result))
	} else if spec, err := names.ParseSpecCollection(pattern); err == nil {
		err2 = core.ListSpecs(ctx, client, spec, filter, generateSpecHandler(&result))
	} else if artifact, err := names.ParseArtifactCollection(pattern); err == nil {
		err2 = core.ListArtifacts(ctx, client, artifact, filter, false, generateArtifactHandler(&result))
	}

	// Then try to match resource names.
	if api, err := names.ParseApi(pattern); err == nil {
		err2 = core.ListAPIs(ctx, client, api, filter, generateApiHandler(&result))
	} else if version, err := names.ParseVersion(pattern); err == nil {
		err2 = core.ListVersions(ctx, client, version, filter, generateVersionHandler(&result))
	} else if spec, err := names.ParseSpec(pattern); err == nil {
		err2 = core.ListSpecs(ctx, client, spec, filter, generateSpecHandler(&result))
	} else if artifact, err := names.ParseArtifact(pattern); err == nil {
		err2 = core.ListArtifacts(ctx, client, artifact, filter, false, generateArtifactHandler(&result))
	}

	if err2 != nil {
		return nil, err2
	}

	return result, nil
}

func generateApiHandler(result *[]ResourceInstance) func(*rpc.Api) {
	return func(api *rpc.Api) {
		apiName, _ := names.ParseApi(api.GetName())
		resource := ApiResource{
			ApiName:         ApiName{Api: apiName},
			UpdateTimestamp: api.UpdateTime.AsTime(),
		}
		(*result) = append((*result), resource)
	}
}

func generateVersionHandler(result *[]ResourceInstance) func(*rpc.ApiVersion) {
	return func(version *rpc.ApiVersion) {
		versionName, _ := names.ParseVersion(version.GetName())
		resource := VersionResource{
			VersionName:     VersionName{Version: versionName},
			UpdateTimestamp: version.UpdateTime.AsTime(),
		}
		(*result) = append((*result), resource)
	}
}

func generateSpecHandler(result *[]ResourceInstance) func(*rpc.ApiSpec) {
	return func(spec *rpc.ApiSpec) {
		specName, _ := names.ParseSpec(spec.GetName())
		resource := SpecResource{
			SpecName:        SpecName{Spec: specName},
			UpdateTimestamp: spec.RevisionUpdateTime.AsTime(),
		}
		(*result) = append((*result), resource)
	}
}

func generateArtifactHandler(result *[]ResourceInstance) func(*rpc.Artifact) {
	return func(artifact *rpc.Artifact) {
		artifactName, _ := names.ParseArtifact(artifact.GetName())
		resource := ArtifactResource{
			ArtifactName:    ArtifactName{Artifact: artifactName},
			UpdateTimestamp: artifact.UpdateTime.AsTime(),
		}
		(*result) = append((*result), resource)
	}
}
