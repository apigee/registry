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

func listResources(ctx context.Context, client connection.Client, pattern, filter string) ([]resourceInstance, error) {
	var result []resourceInstance
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

func generateApiHandler(result *[]resourceInstance) func(*rpc.Api) {
	return func(api *rpc.Api) {
		parsedApiName, err := names.ParseApi(api.GetName())
		if err != nil {
			panic(err)
		}
		resource := apiResource{
			apiName:         apiName{api: parsedApiName},
			updateTimestamp: api.UpdateTime.AsTime(),
		}
		(*result) = append((*result), resource)
	}
}

func generateVersionHandler(result *[]resourceInstance) func(*rpc.ApiVersion) {
	return func(version *rpc.ApiVersion) {
		parsedVersionName, err := names.ParseVersion(version.GetName())
		if err != nil {
			panic(err)
		}
		resource := versionResource{
			versionName:     versionName{version: parsedVersionName},
			updateTimestamp: version.UpdateTime.AsTime(),
		}
		(*result) = append((*result), resource)
	}
}

func generateSpecHandler(result *[]resourceInstance) func(*rpc.ApiSpec) {
	return func(spec *rpc.ApiSpec) {
		parsedSpecName, err := names.ParseSpec(spec.GetName())
		if err != nil {
			panic(err)
		}
		resource := specResource{
			specName:        specName{spec: parsedSpecName},
			updateTimestamp: spec.RevisionUpdateTime.AsTime(),
		}
		(*result) = append((*result), resource)
	}
}

func generateArtifactHandler(result *[]resourceInstance) func(*rpc.Artifact) {
	return func(artifact *rpc.Artifact) {
		parsedArtifactName, err := names.ParseArtifact(artifact.GetName())
		if err != nil {
			panic(err)
		}
		resource := artifactResource{
			artifactName:    artifactName{artifact: parsedArtifactName},
			updateTimestamp: artifact.UpdateTime.AsTime(),
		}
		(*result) = append((*result), resource)
	}
}
