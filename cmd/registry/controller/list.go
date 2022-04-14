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
	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

func listResources(ctx context.Context, client connection.Client, pattern, filter string) ([]patterns.ResourceInstance, error) {
	var result []patterns.ResourceInstance
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

func generateApiHandler(result *[]patterns.ResourceInstance) func(*rpc.Api) error {
	return func(api *rpc.Api) error {
		name, err := names.ParseApi(api.GetName())
		if err != nil {
			return err
		}

		(*result) = append((*result), patterns.ApiResource{
			ApiName:   patterns.ApiName{Name: name},
			Timestamp: api.UpdateTime.AsTime(),
		})

		return nil
	}
}

func generateVersionHandler(result *[]patterns.ResourceInstance) func(*rpc.ApiVersion) error {
	return func(version *rpc.ApiVersion) error {
		name, err := names.ParseVersion(version.GetName())
		if err != nil {
			return err
		}

		(*result) = append((*result), patterns.VersionResource{
			VersionName: patterns.VersionName{Name: name},
			Timestamp:   version.UpdateTime.AsTime(),
		})

		return nil
	}
}

func generateSpecHandler(result *[]patterns.ResourceInstance) func(*rpc.ApiSpec) error {
	return func(spec *rpc.ApiSpec) error {
		name, err := names.ParseSpec(spec.GetName())
		if err != nil {
			return err
		}

		(*result) = append((*result), patterns.SpecResource{
			SpecName:  patterns.SpecName{Name: name},
			Timestamp: spec.RevisionUpdateTime.AsTime(),
		})

		return nil
	}
}

func generateArtifactHandler(result *[]patterns.ResourceInstance) func(*rpc.Artifact) error {
	return func(artifact *rpc.Artifact) error {
		name, err := names.ParseArtifact(artifact.GetName())
		if err != nil {
			return err
		}

		(*result) = append((*result), patterns.ArtifactResource{
			ArtifactName: patterns.ArtifactName{Name: name},
			Timestamp:    artifact.UpdateTime.AsTime(),
		})

		return nil
	}
}
