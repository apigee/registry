package list

import (
	"github.com/apigee/registry/cmd/control_loop/resources"
	"github.com/apigee/registry/rpc"
)

func GenerateArtifactHandler(result *[]resources.Resource) func(*rpc.Artifact) {
	return func(artifact *rpc.Artifact) {
		resource := resources.ArtifactResource{Artifact: artifact}
		(*result) = append((*result), resource)
	}
}


