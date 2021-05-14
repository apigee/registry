package list

import (
	"github.com/apigee/registry/server/names"
	"github.com/apigee/registry/connection"
	"context"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/control_loop/resources"

)

func ListResources(ctx context.Context, client connection.Client, pattern, filter string) ([]resources.Resource, error) {
 	var result []resources.Resource 
 	var err error

	// First try to match collection names.
	if m := names.ApisRegexp().FindStringSubmatch(pattern); m != nil {
		err = core.ListAPIs(ctx, client, m, filter, GenerateApiHandler(&result))
	} else if m := names.SpecsRegexp().FindStringSubmatch(pattern); m != nil {
		err = core.ListSpecs(ctx, client, m, filter, GenerateSpecHandler(&result))
	} else if m := names.ArtifactsRegexp().FindStringSubmatch(pattern); m != nil {
		err = core.ListArtifacts(ctx, client, m, filter, false, GenerateArtifactHandler(&result))
	}

	// Then try to match resource names.
	if m := names.ApiRegexp().FindStringSubmatch(pattern); m != nil {
		err = core.ListAPIs(ctx, client, m, filter, GenerateApiHandler(&result))
	} else if m := names.SpecRegexp().FindStringSubmatch(pattern); m != nil {
		err = core.ListSpecs(ctx, client, m, filter, GenerateSpecHandler(&result))
	} else if m := names.ArtifactRegexp().FindStringSubmatch(pattern); m != nil {
		err = core.ListArtifacts(ctx, client, m, filter, false, GenerateArtifactHandler(&result))
	}

	if err != nil {
		return nil, err
	}

	return result, err
}
 
