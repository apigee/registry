// Copyright 2023 Google LLC.
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
	"context"

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
)

// SubtreeVisitor is a Visitor implementation that can be passed to the
// Visit() function. It will run the handlers in Visitor against each
// resource directly matched in Visit() as well as each resource that is
// a descendant of those resources. (For example, if the pattern matches
// a Version, the SubtreeVisitor will be run against all of that Versions's
// Artifacts and Specs, and the Spec's Artifacts as well.)
// Only the current revisions of Specs and Deployments are visited.
type SubtreeVisitor struct {
	Options VisitorOptions
	Visitor Visitor
}

func (v *SubtreeVisitor) ProjectHandler() ProjectHandler {
	return func(ctx context.Context, message *rpc.Project) error {
		name, _ := names.ParseProject(message.GetName())

		if err := v.Visitor.ProjectHandler()(ctx, message); err != nil {
			return err
		}
		if err := ListArtifacts(ctx, v.Options.RegistryClient, name.Artifact("-"), v.Options.PageSize, v.Options.Filter, v.Options.GetContents, v.ArtifactHandler()); err != nil {
			return err
		}
		if err := ListAPIs(ctx, v.Options.RegistryClient, name.Api("-"), v.Options.PageSize, v.Options.Filter, v.ApiHandler()); err != nil {
			return err
		}
		return nil
	}
}

func (v *SubtreeVisitor) ApiHandler() ApiHandler {
	return func(ctx context.Context, message *rpc.Api) error {
		name, _ := names.ParseApi(message.GetName())

		if err := v.Visitor.ApiHandler()(ctx, message); err != nil {
			return err
		}
		if err := ListArtifacts(ctx, v.Options.RegistryClient, name.Artifact("-"), v.Options.PageSize, v.Options.Filter, v.Options.GetContents, v.ArtifactHandler()); err != nil {
			return err
		}
		if err := ListVersions(ctx, v.Options.RegistryClient, name.Version("-"), v.Options.PageSize, v.Options.Filter, v.VersionHandler()); err != nil {
			return err
		}
		if err := ListDeployments(ctx, v.Options.RegistryClient, name.Deployment("-"), v.Options.PageSize, v.Options.Filter, v.DeploymentHandler()); err != nil {
			return err
		}
		return nil
	}
}

func (v *SubtreeVisitor) VersionHandler() VersionHandler {
	return func(ctx context.Context, message *rpc.ApiVersion) error {
		name, _ := names.ParseVersion(message.GetName())

		if err := v.Visitor.VersionHandler()(ctx, message); err != nil {
			return err
		}
		if err := ListArtifacts(ctx, v.Options.RegistryClient, name.Artifact("-"), v.Options.PageSize, v.Options.Filter, v.Options.GetContents, v.ArtifactHandler()); err != nil {
			return err
		}
		if err := ListSpecs(ctx, v.Options.RegistryClient, name.Spec("-"), v.Options.PageSize, v.Options.Filter, v.Options.GetContents, v.SpecHandler()); err != nil {
			return err
		}
		return nil
	}
}

func (v *SubtreeVisitor) DeploymentHandler() DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		name, _ := names.ParseDeployment(message.GetName())

		if err := v.Visitor.DeploymentHandler()(ctx, message); err != nil {
			return err
		}
		if err := ListArtifacts(ctx, v.Options.RegistryClient, name.Artifact("-"), v.Options.PageSize, v.Options.Filter, v.Options.GetContents, v.ArtifactHandler()); err != nil {
			return err
		}
		return nil
	}
}

func (v *SubtreeVisitor) DeploymentRevisionHandler() DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		name, _ := names.ParseDeploymentRevision(message.GetName())

		if err := v.Visitor.DeploymentRevisionHandler()(ctx, message); err != nil {
			return err
		}
		if err := ListArtifacts(ctx, v.Options.RegistryClient, name.Artifact("-"), v.Options.PageSize, v.Options.Filter, v.Options.GetContents, v.ArtifactHandler()); err != nil {
			return err
		}
		return nil
	}
}

func (v *SubtreeVisitor) SpecHandler() SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		name, _ := names.ParseSpec(message.GetName())

		if err := v.Visitor.SpecHandler()(ctx, message); err != nil {
			return err
		}
		if err := ListArtifacts(ctx, v.Options.RegistryClient, name.Artifact("-"), v.Options.PageSize, v.Options.Filter, v.Options.GetContents, v.ArtifactHandler()); err != nil {
			return err
		}
		return nil
	}
}

func (v *SubtreeVisitor) SpecRevisionHandler() SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		name, _ := names.ParseSpecRevision(message.GetName())

		if err := v.Visitor.SpecRevisionHandler()(ctx, message); err != nil {
			return err
		}
		if err := ListArtifacts(ctx, v.Options.RegistryClient, name.Artifact("-"), v.Options.PageSize, v.Options.Filter, v.Options.GetContents, v.ArtifactHandler()); err != nil {
			return err
		}
		return nil
	}
}

func (v *SubtreeVisitor) ArtifactHandler() ArtifactHandler {
	return func(ctx context.Context, message *rpc.Artifact) error {
		return v.Visitor.ArtifactHandler()(ctx, message)
	}
}
