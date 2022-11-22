package controller

import (
	"context"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

type listingClient interface {
	ListAPIs(context.Context, names.Api, string, core.ApiHandler) error
	ListVersions(context.Context, names.Version, string, core.VersionHandler) error
	ListSpecs(context.Context, names.Spec, string, core.SpecHandler) error
	ListArtifacts(context.Context, names.Artifact, string, bool, core.ArtifactHandler) error
}

type RegistryLister struct {
	RegistryClient connection.RegistryClient
}

func (r *RegistryLister) ListAPIs(ctx context.Context, api names.Api, filter string, handler core.ApiHandler) error {
	return core.ListAPIs(ctx, r.RegistryClient, api, filter, handler)
}

func (r *RegistryLister) ListVersions(ctx context.Context, version names.Version, filter string, handler core.VersionHandler) error {
	return core.ListVersions(ctx, r.RegistryClient, version, filter, handler)
}

func (r *RegistryLister) ListSpecs(ctx context.Context, spec names.Spec, filter string, handler core.SpecHandler) error {
	return core.ListSpecs(ctx, r.RegistryClient, spec, filter, handler)
}

func (r *RegistryLister) ListArtifacts(ctx context.Context, artifact names.Artifact, filter string, contents bool, handler core.ArtifactHandler) error {
	return core.ListArtifacts(ctx, r.RegistryClient, artifact, filter, contents, handler)
}

func listResources(ctx context.Context, client listingClient, pattern, filter string) ([]patterns.ResourceInstance, error) {
	var result []patterns.ResourceInstance
	var err2 error

	// First try to match collection names.
	if api, err := names.ParseApiCollection(pattern); err == nil {
		err2 = client.ListAPIs(ctx, api, filter, generateApiHandler(&result))
	} else if version, err := names.ParseVersionCollection(pattern); err == nil {
		err2 = client.ListVersions(ctx, version, filter, generateVersionHandler(&result))
	} else if spec, err := names.ParseSpecCollection(pattern); err == nil {
		err2 = client.ListSpecs(ctx, spec, filter, generateSpecHandler(&result))
	} else if artifact, err := names.ParseArtifactCollection(pattern); err == nil {
		err2 = client.ListArtifacts(ctx, artifact, filter, true, generateArtifactHandler(&result))
	}

	// Then try to match resource names.
	if api, err := names.ParseApi(pattern); err == nil {
		err2 = client.ListAPIs(ctx, api, filter, generateApiHandler(&result))
	} else if version, err := names.ParseVersion(pattern); err == nil {
		err2 = client.ListVersions(ctx, version, filter, generateVersionHandler(&result))
	} else if spec, err := names.ParseSpec(pattern); err == nil {
		err2 = client.ListSpecs(ctx, spec, filter, generateSpecHandler(&result))
	} else if artifact, err := names.ParseArtifact(pattern); err == nil {
		err2 = client.ListArtifacts(ctx, artifact, filter, true, generateArtifactHandler(&result))
	}

	if err2 != nil {
		return nil, err2
	}
	return result, nil
}

func generateApiHandler(result *[]patterns.ResourceInstance) func(*rpc.Api) error {
	return func(api *rpc.Api) error {
		(*result) = append((*result), patterns.ApiResource{
			Api: api,
		})

		return nil
	}
}

func generateVersionHandler(result *[]patterns.ResourceInstance) func(*rpc.ApiVersion) error {
	return func(version *rpc.ApiVersion) error {
		(*result) = append((*result), patterns.VersionResource{
			Version: version,
		})

		return nil
	}
}

func generateSpecHandler(result *[]patterns.ResourceInstance) func(*rpc.ApiSpec) error {
	return func(spec *rpc.ApiSpec) error {
		(*result) = append((*result), patterns.SpecResource{
			Spec: spec,
		})

		return nil
	}
}

func generateArtifactHandler(result *[]patterns.ResourceInstance) func(*rpc.Artifact) error {
	return func(artifact *rpc.Artifact) error {
		(*result) = append((*result), patterns.ArtifactResource{
			Artifact: artifact,
		})

		return nil
	}
}
