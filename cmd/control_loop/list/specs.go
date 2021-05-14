package list

import (
	"github.com/apigee/registry/cmd/control_loop/resources"
	"github.com/apigee/registry/rpc"
)

func GenerateSpecHandler(result *[]resources.Resource) func(*rpc.ApiSpec) {
	return func(spec *rpc.ApiSpec) {
		resource := resources.SpecResource{Spec: spec}
		(*result) = append((*result), resource)
	}
}


