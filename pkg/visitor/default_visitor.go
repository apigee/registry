// Copyright 2023 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package visitor

import (
	"fmt"

	"github.com/apigee/registry/rpc"
)

// DefaultVisitor provides default handlers for all types that return "unsupported" errors.
// Include it as an extension to fill out incomplete visitor implementations.
type DefaultVisitor struct {
}

func (v *DefaultVisitor) ProjectHandler() ProjectHandler {
	return func(message *rpc.Project) error {
		return fmt.Errorf("unsupported operand type: %T", message)
	}
}

func (v *DefaultVisitor) ApiHandler() ApiHandler {
	return func(message *rpc.Api) error {
		return fmt.Errorf("unsupported operand type: %T", message)
	}
}

func (v *DefaultVisitor) VersionHandler() VersionHandler {
	return func(message *rpc.ApiVersion) error {
		return fmt.Errorf("unsupported operand type: %T", message)
	}
}

func (v *DefaultVisitor) DeploymentHandler() DeploymentHandler {
	return func(message *rpc.ApiDeployment) error {
		return fmt.Errorf("unsupported operand type: %T", message)
	}
}

func (v *DefaultVisitor) DeploymentRevisionHandler() DeploymentHandler {
	return func(message *rpc.ApiDeployment) error {
		return fmt.Errorf("unsupported operand type: %T", message)
	}
}

func (v *DefaultVisitor) SpecHandler() SpecHandler {
	return func(message *rpc.ApiSpec) error {
		return fmt.Errorf("unsupported operand type: %T", message)
	}
}

func (v *DefaultVisitor) SpecRevisionHandler() SpecHandler {
	return func(message *rpc.ApiSpec) error {
		return fmt.Errorf("unsupported operand type: %T", message)
	}
}

func (v *DefaultVisitor) ArtifactHandler() ArtifactHandler {
	return func(message *rpc.Artifact) error {
		return fmt.Errorf("unsupported operand type: %T", message)
	}
}
