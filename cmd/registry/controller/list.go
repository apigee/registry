package controller

import (
	"context"

	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
)

type listingClient interface {
	ListAPIs(context.Context, names.Api, string, visitor.ApiHandler) error
	ListVersions(context.Context, names.Version, string, visitor.VersionHandler) error
	ListSpecs(context.Context, names.Spec, string, visitor.SpecHandler) error
	ListArtifacts(context.Context, names.Artifact, string, bool, visitor.ArtifactHandler) error
}

type RegistryLister struct {
	RegistryClient connection.RegistryClient
}

func (r *RegistryLister) ListAPIs(ctx context.Context, api names.Api, filter string, handler visitor.ApiHandler) error {
	return visitor.ListAPIs(ctx, r.RegistryClient, api, 0, filter, handler)
}

func (r *RegistryLister) ListVersions(ctx context.Context, version names.Version, filter string, handler visitor.VersionHandler) error {
	return visitor.ListVersions(ctx, r.RegistryClient, version, 0, filter, handler)
}

func (r *RegistryLister) ListSpecs(ctx context.Context, spec names.Spec, filter string, handler visitor.SpecHandler) error {
	return visitor.ListSpecs(ctx, r.RegistryClient, spec, 0, filter, false, handler)
}

func (r *RegistryLister) ListArtifacts(ctx context.Context, artifact names.Artifact, filter string, contents bool, handler visitor.ArtifactHandler) error {
	return visitor.ListArtifacts(ctx, r.RegistryClient, artifact, 0, filter, contents, handler)
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

func generateApiHandler(result *[]patterns.ResourceInstance) func(context.Context, *rpc.Api) error {
	return func(ctx context.Context, api *rpc.Api) error {
		(*result) = append((*result), patterns.ApiResource{
			Api: api,
		})

		return nil
	}
}

func generateVersionHandler(result *[]patterns.ResourceInstance) func(context.Context, *rpc.ApiVersion) error {
	return func(ctx context.Context, version *rpc.ApiVersion) error {
		(*result) = append((*result), patterns.VersionResource{
			Version: version,
		})

		return nil
	}
}

func generateSpecHandler(result *[]patterns.ResourceInstance) func(context.Context, *rpc.ApiSpec) error {
	return func(ctx context.Context, spec *rpc.ApiSpec) error {
		(*result) = append((*result), patterns.SpecResource{
			Spec: spec,
		})

		return nil
	}
}

func generateArtifactHandler(result *[]patterns.ResourceInstance) func(context.Context, *rpc.Artifact) error {
	return func(ctx context.Context, artifact *rpc.Artifact) error {
		(*result) = append((*result), patterns.ArtifactResource{
			Artifact: artifact,
		})

		return nil
	}
}
