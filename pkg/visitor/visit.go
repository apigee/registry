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
	"errors"
	"fmt"
	"strings"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
)

type Visitor interface {
	ProjectHandler() ProjectHandler
	ApiHandler() ApiHandler
	DeploymentHandler() DeploymentHandler
	DeploymentRevisionHandler() DeploymentHandler
	VersionHandler() VersionHandler
	SpecHandler() SpecHandler
	SpecRevisionHandler() SpecHandler
	ArtifactHandler() ArtifactHandler
}

type VisitorOptions struct {
	RegistryClient  connection.RegistryClient
	AdminClient     connection.AdminClient
	Pattern         string
	PageSize        int32
	Filter          string
	GetContents     bool
	ImplicitProject *rpc.Project // used as placeholder if Project is unaccessible
}

// Visit traverses a registry, applying the Visitor to each selected resource.
func Visit(ctx context.Context, v Visitor, options VisitorOptions) error {
	filter := options.Filter
	name := options.Pattern
	ac := options.AdminClient
	rc := options.RegistryClient
	pageSize := options.PageSize

	// First try to match collection names.
	if project, err := names.ParseProjectCollection(name); err == nil {
		return ListProjects(ctx, ac, project, options.ImplicitProject, pageSize, filter, v.ProjectHandler())
	} else if api, err := names.ParseApiCollection(name); err == nil {
		return ListAPIs(ctx, rc, api, pageSize, filter, v.ApiHandler())
	} else if deployment, err := names.ParseDeploymentCollection(name); err == nil {
		return ListDeployments(ctx, rc, deployment, pageSize, filter, v.DeploymentHandler())
	} else if rev, err := names.ParseDeploymentRevisionCollection(name); err == nil {
		return ListDeploymentRevisions(ctx, rc, rev, pageSize, filter, v.DeploymentRevisionHandler())
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return ListVersions(ctx, rc, version, pageSize, filter, v.VersionHandler())
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return ListSpecs(ctx, rc, spec, pageSize, filter, options.GetContents, v.SpecHandler())
	} else if rev, err := names.ParseSpecRevisionCollection(name); err == nil {
		return ListSpecRevisions(ctx, rc, rev, pageSize, filter, options.GetContents, v.SpecRevisionHandler())
	} else if artifact, err := names.ParseArtifactCollection(name); err == nil {
		return ListArtifacts(ctx, rc, artifact, pageSize, filter, options.GetContents, v.ArtifactHandler())
	}
	// Then try to match resource names containing wildcards, these also are treated as collections.
	if strings.Contains(name, "/-") || strings.Contains(name, "@-") {
		if project, err := names.ParseProject(name); err == nil {
			return ListProjects(ctx, ac, project, options.ImplicitProject, pageSize, filter, v.ProjectHandler())
		} else if api, err := names.ParseApi(name); err == nil {
			return ListAPIs(ctx, rc, api, pageSize, filter, v.ApiHandler())
		} else if deployment, err := names.ParseDeployment(name); err == nil {
			return ListDeployments(ctx, rc, deployment, pageSize, filter, v.DeploymentHandler())
		} else if rev, err := names.ParseDeploymentRevision(name); err == nil {
			return ListDeploymentRevisions(ctx, rc, rev, pageSize, filter, v.DeploymentRevisionHandler())
		} else if version, err := names.ParseVersion(name); err == nil {
			return ListVersions(ctx, rc, version, pageSize, filter, v.VersionHandler())
		} else if spec, err := names.ParseSpec(name); err == nil {
			return ListSpecs(ctx, rc, spec, pageSize, filter, options.GetContents, v.SpecHandler())
		} else if rev, err := names.ParseSpecRevision(name); err == nil {
			return ListSpecRevisions(ctx, rc, rev, pageSize, filter, options.GetContents, v.SpecRevisionHandler())
		} else if artifact, err := names.ParseArtifact(name); err == nil {
			return ListArtifacts(ctx, rc, artifact, pageSize, filter, options.GetContents, v.ArtifactHandler())
		}
		return fmt.Errorf("unsupported pattern %+v", name)
	}
	// If we get here, name designates an individual resource to be displayed.
	// So if a filter was specified, that's an error.
	if filter != "" {
		return errors.New("--filter must not be specified for a non-collection resource")
	}
	// Finally, match individual resources
	if project, err := names.ParseProject(name); err == nil {
		return GetProject(ctx, ac, project, options.ImplicitProject, v.ProjectHandler())
	} else if api, err := names.ParseApi(name); err == nil {
		return GetAPI(ctx, rc, api, v.ApiHandler())
	} else if deployment, err := names.ParseDeployment(name); err == nil {
		return GetDeployment(ctx, rc, deployment, v.DeploymentHandler())
	} else if deployment, err := names.ParseDeploymentRevision(name); err == nil {
		return GetDeploymentRevision(ctx, rc, deployment, v.DeploymentRevisionHandler())
	} else if version, err := names.ParseVersion(name); err == nil {
		return GetVersion(ctx, rc, version, v.VersionHandler())
	} else if spec, err := names.ParseSpec(name); err == nil {
		return GetSpec(ctx, rc, spec, options.GetContents, v.SpecHandler())
	} else if spec, err := names.ParseSpecRevision(name); err == nil {
		return GetSpecRevision(ctx, rc, spec, options.GetContents, v.SpecRevisionHandler())
	} else if artifact, err := names.ParseArtifact(name); err == nil {
		return GetArtifact(ctx, rc, artifact, options.GetContents, v.ArtifactHandler())
	} else {
		return fmt.Errorf("unsupported pattern %+v", name)
	}
}
