package scoring

import (
	"context"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

type artifactClient interface {
	GetArtifact(context.Context, names.Artifact, bool, core.ArtifactHandler) error
	SetArtifact(context.Context, *rpc.Artifact) error
	ListArtifacts(context.Context, names.Artifact, string, bool, core.ArtifactHandler) error
}

type RegistryArtifactClient struct {
	RegistryClient connection.RegistryClient
}

func (r *RegistryArtifactClient) GetArtifact(ctx context.Context, artifact names.Artifact, getContents bool, handler core.ArtifactHandler) error {
	return core.GetArtifact(ctx, r.RegistryClient, artifact, getContents, handler)
}

func (r *RegistryArtifactClient) SetArtifact(ctx context.Context, artifact *rpc.Artifact) error {
	return core.SetArtifact(ctx, r.RegistryClient, artifact)
}

func (r *RegistryArtifactClient) ListArtifacts(ctx context.Context, artifact names.Artifact, filter string, contents bool, handler core.ArtifactHandler) error {
	return core.ListArtifacts(ctx, r.RegistryClient, artifact, filter, contents, handler)
}
