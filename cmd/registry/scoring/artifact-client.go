package scoring

import (
	"context"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
)

type artifactClient interface {
	GetArtifact(context.Context, names.Artifact, bool, visitor.ArtifactHandler) error
	SetArtifact(context.Context, *rpc.Artifact) error
	ListArtifacts(context.Context, names.Artifact, string, bool, visitor.ArtifactHandler) error
}

type RegistryArtifactClient struct {
	RegistryClient connection.RegistryClient
}

func (r *RegistryArtifactClient) GetArtifact(ctx context.Context, artifact names.Artifact, getContents bool, handler visitor.ArtifactHandler) error {
	return visitor.GetArtifact(ctx, r.RegistryClient, artifact, getContents, handler)
}

func (r *RegistryArtifactClient) SetArtifact(ctx context.Context, artifact *rpc.Artifact) error {
	return visitor.SetArtifact(ctx, r.RegistryClient, artifact)
}

func (r *RegistryArtifactClient) ListArtifacts(ctx context.Context, artifact names.Artifact, filter string, contents bool, handler visitor.ArtifactHandler) error {
	return visitor.ListArtifacts(ctx, r.RegistryClient, artifact, 0, filter, contents, handler)
}
