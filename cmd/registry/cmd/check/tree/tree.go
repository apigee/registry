// Copyright 2022 Google LLC. All Rights Reserved.
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

package tree

import (
	"context"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/server/registry/names"
)

// HandlerSet should return an appropriate Handler for each
// rpc.* type.
type HandlerSet interface {
	ProjectHandler() visitor.ProjectHandler
	ApiHandler() visitor.ApiHandler
	DeploymentHandler() visitor.DeploymentHandler
	VersionHandler() visitor.VersionHandler
	SpecHandler() visitor.SpecHandler
	ArtifactHandler() visitor.ArtifactHandler
}

// ListSubresources calls List* on all subtypes below the supplied root name.
// Specific revisions are not listed unless they are specified for the root name.
// The passed filter is given to each List* function, regardless of type.
// The passed HandlerSet is not checked for validity and must include all types
// of handlers that may be called beneath the root.
func ListSubresources(ctx context.Context,
	adminClient connection.AdminClient,
	client connection.RegistryClient,
	root names.Name,
	filter string,
	getContents bool,
	handler HandlerSet) (err error) {
	switch name := root.(type) {
	case names.Project:
		err = visitor.ListProjects(ctx, adminClient, name, true, filter, handler.ProjectHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
		err = visitor.ListAPIs(ctx, client, name.Api("-"), filter, handler.ApiHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Api("-").Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
		err = visitor.ListDeployments(ctx, client, name.Api("-").Deployment("-"), filter, handler.DeploymentHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Api("-").Deployment("-").Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
		err = visitor.ListVersions(ctx, client, name.Api("-").Version("-"), filter, handler.VersionHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Api("-").Version("-").Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
		err = visitor.ListSpecs(ctx, client, name.Api("-").Version("-").Spec("-"), filter, getContents, handler.SpecHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Api("-").Version("-").Spec("-").Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
	case names.Api:
		err = visitor.ListAPIs(ctx, client, name, filter, handler.ApiHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
		err = visitor.ListDeployments(ctx, client, name.Deployment("-"), filter, handler.DeploymentHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Deployment("-").Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
		err = visitor.ListVersions(ctx, client, name.Version("-"), filter, handler.VersionHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Version("-").Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
		err = visitor.ListSpecs(ctx, client, name.Version("-").Spec("-"), filter, getContents, handler.SpecHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Version("-").Spec("-").Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
	case names.Deployment:
		err = visitor.ListDeployments(ctx, client, name, filter, handler.DeploymentHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
	case names.DeploymentRevision:
		err = visitor.ListDeploymentRevisions(ctx, client, name, filter, handler.DeploymentHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
	case names.Version:
		err = visitor.ListVersions(ctx, client, name, filter, handler.VersionHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
		err = visitor.ListSpecs(ctx, client, name.Spec("-"), filter, getContents, handler.SpecHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Spec("-").Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
	case names.Spec:
		err = visitor.ListSpecs(ctx, client, name, filter, getContents, handler.SpecHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
	case names.SpecRevision:
		err = visitor.ListSpecRevisions(ctx, client, name, filter, getContents, handler.SpecHandler())
		if err != nil {
			return err
		}
		err = visitor.ListArtifacts(ctx, client, name.Artifact("-"), filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
	case names.Artifact:
		err = visitor.ListArtifacts(ctx, client, name, filter, getContents, handler.ArtifactHandler())
		if err != nil {
			return err
		}
	}
	return nil
}
