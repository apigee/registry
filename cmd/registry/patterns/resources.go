// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package patterns

import (
	"context"
	"time"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
)

const ResourceUpdateThreshold = time.Second * 2

// This interface is used to describe generic resource names
// Example:
// projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi
// projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-
type ResourceName interface {
	Artifact() string
	Spec() string
	Version() string
	Api() string
	Project() string
	String() string
	ParentName() ResourceName
}

type SpecName struct {
	Name       names.Spec
	RevisionID string
}

func (s SpecName) Artifact() string {
	return ""
}

func (s SpecName) Spec() string {
	return s.String()
}

func (s SpecName) Version() string {
	return s.Name.Version().String()
}

func (s SpecName) Api() string {
	return s.Name.Api().String()
}

func (s SpecName) Project() string {
	return s.Name.Project().String()
}

func (s SpecName) String() string {
	if s.RevisionID == "" {
		return s.Name.String()
	} else {
		return s.Name.String() + "@" + s.RevisionID
	}
}

func (s SpecName) ParentName() ResourceName {
	// Validate the parent name and return
	if version, err := names.ParseVersion(s.Name.Parent()); err == nil {
		return VersionName{
			Name: version,
		}
	} else if version, err := names.ParseVersionCollection(s.Name.Parent()); err == nil {
		return VersionName{
			Name: version,
		}
	}
	return VersionName{}
}

type VersionName struct {
	Name names.Version
}

func (v VersionName) Artifact() string {
	return ""
}

func (v VersionName) Spec() string {
	return ""
}

func (v VersionName) Version() string {
	return v.Name.String()
}

func (v VersionName) Api() string {
	return v.Name.Api().String()
}

func (v VersionName) Project() string {
	return v.Name.Project().String()
}

func (v VersionName) String() string {
	return v.Name.String()
}

func (v VersionName) ParentName() ResourceName {
	// Validate the parent name and return
	if api, err := names.ParseApi(v.Name.Parent()); err == nil {
		return ApiName{
			Name: api,
		}
	} else if api, err := names.ParseApiCollection(v.Name.Parent()); err == nil {
		return ApiName{
			Name: api,
		}
	}

	return ApiName{}
}

type ApiName struct {
	Name names.Api
}

func (a ApiName) Artifact() string {
	return ""
}

func (a ApiName) Spec() string {
	return ""
}

func (a ApiName) Version() string {
	return ""
}

func (a ApiName) Api() string {
	return a.Name.String()
}

func (a ApiName) Project() string {
	return a.Name.Project().String()
}

func (a ApiName) String() string {
	return a.Name.String()
}

func (a ApiName) ParentName() ResourceName {
	// Validate the parent name and return
	if project, err := names.ParseProject(a.Name.Parent()); err == nil {
		return ProjectName{
			Name: project,
		}
	} else if project, err := names.ParseProjectWithLocation(a.Name.Parent()); err == nil {
		return ProjectName{
			Name: project,
		}
	} else if project, err := names.ParseProjectCollection(a.Name.Parent()); err == nil {
		return ProjectName{
			Name: project,
		}
	}

	return ProjectName{}
}

type ProjectName struct {
	Name names.Project
}

func (p ProjectName) Artifact() string {
	return ""
}

func (p ProjectName) Spec() string {
	return ""
}

func (p ProjectName) Version() string {
	return ""
}

func (p ProjectName) Api() string {
	return ""
}

func (p ProjectName) Project() string {
	return p.Name.String()
}

func (p ProjectName) String() string {
	return p.Name.String()
}

func (p ProjectName) ParentName() ResourceName {
	return nil
}

type ArtifactName struct {
	Name names.Artifact
}

func (ar ArtifactName) Artifact() string {
	return ar.Name.String()
}

func (ar ArtifactName) Spec() string {
	specPattern := names.SpecRevision{
		ProjectID:  ar.Name.ProjectID(),
		ApiID:      ar.Name.ApiID(),
		VersionID:  ar.Name.VersionID(),
		SpecID:     ar.Name.SpecID(),
		RevisionID: ar.Name.RevisionID(),
	}

	// Validate the generated name
	if spec, err := names.ParseSpecRevision(specPattern.String()); err == nil {
		return spec.String()
	} else if _, err := names.ParseSpecRevisionCollection(specPattern.String()); err == nil {
		return spec.String()
	}

	return ""
}

func (ar ArtifactName) Version() string {
	versionPattern := names.Version{
		ProjectID: ar.Name.ProjectID(),
		ApiID:     ar.Name.ApiID(),
		VersionID: ar.Name.VersionID(),
	}
	// Validate the generated name
	if version, err := names.ParseVersion(versionPattern.String()); err == nil {
		return version.String()
	} else if _, err := names.ParseVersionCollection(versionPattern.String()); err == nil {
		return version.String()
	}

	return ""
}

func (ar ArtifactName) Api() string {
	apiPattern := names.Api{
		ProjectID: ar.Name.ProjectID(),
		ApiID:     ar.Name.ApiID(),
	}
	// Validate the generated name
	if _, err := names.ParseApi(apiPattern.String()); err == nil {
		return apiPattern.String()
	} else if _, err := names.ParseApiCollection(apiPattern.String()); err == nil {
		return apiPattern.String()
	}

	return ""
}

func (ar ArtifactName) Project() string {
	projectPattern := names.Project{
		ProjectID: ar.Name.ProjectID(),
	}
	// Validate the generated name
	if _, err := names.ParseProject(projectPattern.String()); err == nil {
		return projectPattern.String()
	} else if _, err := names.ParseProjectCollection(projectPattern.String()); err == nil {
		return projectPattern.String()
	}

	return ""
}

func (ar ArtifactName) String() string {
	return ar.Name.String()
}

func (ar ArtifactName) ParentName() ResourceName {
	parent := ar.Name.Parent()
	// First try to match collection names.
	if project, err := names.ParseProjectCollection(parent); err == nil {
		return ProjectName{
			Name: project,
		}
	} else if api, err := names.ParseApiCollection(parent); err == nil {
		return ApiName{
			Name: api,
		}
	} else if version, err := names.ParseVersionCollection(parent); err == nil {
		return VersionName{
			Name: version,
		}
	} else if spec, err := names.ParseSpecRevisionCollection(parent); err == nil {
		return SpecName{
			Name:       spec.Spec(),
			RevisionID: spec.RevisionID,
		}
	}

	// Then try to match resource names.
	if project, err := names.ParseProject(parent); err == nil {
		return ProjectName{
			Name: project,
		}
	} else if project, err := names.ParseProjectWithLocation(parent); err == nil {
		return ProjectName{
			Name: project,
		}
	} else if api, err := names.ParseApi(parent); err == nil {
		return ApiName{
			Name: api,
		}
	} else if version, err := names.ParseVersion(parent); err == nil {
		return VersionName{
			Name: version,
		}
	} else if spec, err := names.ParseSpecRevision(parent); err == nil {
		return SpecName{
			Name:       spec.Spec(),
			RevisionID: spec.RevisionID,
		}
	}

	return nil
}

// This interface is used to describe generic resource instances
// ResourceName is embedded, the only additional field is the UpdateTimestamp
type ResourceInstance interface {
	ResourceName() ResourceName
	UpdateTimestamp() time.Time
}

type SpecResource struct {
	Spec *rpc.ApiSpec
}

func (s SpecResource) UpdateTimestamp() time.Time {
	return s.Spec.RevisionUpdateTime.AsTime()
}

func (s SpecResource) ResourceName() ResourceName {
	name, err := names.ParseSpecRevision(s.Spec.GetName())
	if err != nil {
		return nil
	}
	return SpecName{
		Name:       name.Spec(),
		RevisionID: s.Spec.RevisionId,
	}
}

type VersionResource struct {
	Version *rpc.ApiVersion
}

func (v VersionResource) UpdateTimestamp() time.Time {
	return v.Version.UpdateTime.AsTime()
}

func (v VersionResource) ResourceName() ResourceName {
	name, err := names.ParseVersion(v.Version.GetName())
	if err != nil {
		return nil
	}
	return VersionName{Name: name}
}

type ApiResource struct {
	Api *rpc.Api
}

func (a ApiResource) UpdateTimestamp() time.Time {
	return a.Api.UpdateTime.AsTime()
}

func (a ApiResource) ResourceName() ResourceName {
	name, err := names.ParseApi(a.Api.GetName())
	if err != nil {
		return nil
	}
	return ApiName{Name: name}
}

// Project is a special resource which is not available through the registry client.
// Hence we won't store the actual instance of the rpc.Project but instead only store the project name.
// ProjectResource is mainly used to identify by name when we derive the parents of certain artifacts.
type ProjectResource struct {
	ProjectName string
}

func (p ProjectResource) UpdateTimestamp() time.Time {
	return time.Time{}
}

func (p ProjectResource) ResourceName() ResourceName {
	name, err := names.ParseProject(p.ProjectName)
	if err != nil {
		return nil
	}
	return ProjectName{Name: name}
}

type ArtifactResource struct {
	Artifact *rpc.Artifact
}

func (ar ArtifactResource) UpdateTimestamp() time.Time {
	return ar.Artifact.UpdateTime.AsTime()
}

func (ar ArtifactResource) ResourceName() ResourceName {
	name, err := names.ParseArtifact(ar.Artifact.GetName())
	if err != nil {
		return nil
	}
	return ArtifactName{Name: name}
}

func ListResources(ctx context.Context, client connection.RegistryClient, pattern, filter string) ([]ResourceInstance, error) {
	var result []ResourceInstance
	var err2 error

	// First try to match collection names.
	if api, err := names.ParseApiCollection(pattern); err == nil {
		err2 = visitor.ListAPIs(ctx, client, api, 0, filter, generateApiHandler(&result))
	} else if version, err := names.ParseVersionCollection(pattern); err == nil {
		err2 = visitor.ListVersions(ctx, client, version, 0, filter, generateVersionHandler(&result))
	} else if spec, err := names.ParseSpecCollection(pattern); err == nil {
		err2 = visitor.ListSpecs(ctx, client, spec, 0, filter, false, generateSpecHandler(&result))
	} else if rev, err := names.ParseSpecRevisionCollection(pattern); err == nil {
		err2 = visitor.ListSpecRevisions(ctx, client, rev, 0, filter, false, generateSpecHandler(&result))
	} else if artifact, err := names.ParseArtifactCollection(pattern); err == nil {
		err2 = visitor.ListArtifacts(ctx, client, artifact, 0, filter, true, generateArtifactHandler(&result))
	}

	// Then try to match resource names.
	if api, err := names.ParseApi(pattern); err == nil {
		err2 = visitor.ListAPIs(ctx, client, api, 0, filter, generateApiHandler(&result))
	} else if version, err := names.ParseVersion(pattern); err == nil {
		err2 = visitor.ListVersions(ctx, client, version, 0, filter, generateVersionHandler(&result))
	} else if spec, err := names.ParseSpec(pattern); err == nil {
		err2 = visitor.ListSpecs(ctx, client, spec, 0, filter, false, generateSpecHandler(&result))
	} else if rev, err := names.ParseSpecRevision(pattern); err == nil {
		err2 = visitor.ListSpecRevisions(ctx, client, rev, 0, filter, false, generateSpecHandler(&result))
	} else if artifact, err := names.ParseArtifact(pattern); err == nil {
		err2 = visitor.ListArtifacts(ctx, client, artifact, 0, filter, true, generateArtifactHandler(&result))
	}

	if err2 != nil {
		return nil, err2
	}

	return result, nil
}

func generateApiHandler(result *[]ResourceInstance) func(context.Context, *rpc.Api) error {
	return func(ctx context.Context, api *rpc.Api) error {
		(*result) = append((*result), ApiResource{
			Api: api,
		})
		return nil
	}
}

func generateVersionHandler(result *[]ResourceInstance) func(context.Context, *rpc.ApiVersion) error {
	return func(ctx context.Context, version *rpc.ApiVersion) error {
		(*result) = append((*result), VersionResource{
			Version: version,
		})
		return nil
	}
}

func generateSpecHandler(result *[]ResourceInstance) func(context.Context, *rpc.ApiSpec) error {
	return func(ctx context.Context, spec *rpc.ApiSpec) error {
		(*result) = append((*result), SpecResource{
			Spec: spec,
		})
		return nil
	}
}

func generateArtifactHandler(result *[]ResourceInstance) func(context.Context, *rpc.Artifact) error {
	return func(ctx context.Context, artifact *rpc.Artifact) error {
		(*result) = append((*result), ArtifactResource{
			Artifact: artifact,
		})
		return nil
	}
}
