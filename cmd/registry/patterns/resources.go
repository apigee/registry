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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

// This interface is used to describe generic resource names
// Example:
// projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml
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
	Name names.Spec
}

func (s SpecName) Artifact() string {
	return ""
}

func (s SpecName) Spec() string {
	return s.Name.String()
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
	return s.Name.String()
}

func (s SpecName) ParentName() ResourceName {
	// Validate the parent name aand return
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
	// Validate the parent name aand return
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
	// Validate the parent name aand return
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
	specPattern := names.Spec{
		ProjectID: ar.Name.ProjectID(),
		ApiID:     ar.Name.ApiID(),
		VersionID: ar.Name.VersionID(),
		SpecID:    ar.Name.SpecID(),
	}

	// Validate the generated name
	if spec, err := names.ParseSpec(specPattern.String()); err == nil {
		return spec.String()
	} else if _, err := names.ParseSpecCollection(specPattern.String()); err == nil {
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
	} else if spec, err := names.ParseSpecCollection(parent); err == nil {
		return SpecName{
			Name: spec,
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
	} else if spec, err := names.ParseSpec(parent); err == nil {
		return SpecName{
			Name: spec,
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
	SpecName  ResourceName
	Timestamp time.Time
}

func (s SpecResource) UpdateTimestamp() time.Time {
	return s.Timestamp
}

func (s SpecResource) ResourceName() ResourceName {
	return s.SpecName
}

type VersionResource struct {
	VersionName ResourceName
	Timestamp   time.Time
}

func (v VersionResource) UpdateTimestamp() time.Time {
	return v.Timestamp
}

func (v VersionResource) ResourceName() ResourceName {
	return v.VersionName
}

type ApiResource struct {
	ApiName   ResourceName
	Timestamp time.Time
}

func (a ApiResource) UpdateTimestamp() time.Time {
	return a.Timestamp
}

func (a ApiResource) ResourceName() ResourceName {
	return a.ApiName
}

type ProjectResource struct {
	ProjectName ResourceName
	Timestamp   time.Time
}

func (p ProjectResource) UpdateTimestamp() time.Time {
	return p.Timestamp
}

func (p ProjectResource) ResourceName() ResourceName {
	return p.ProjectName
}

type ArtifactResource struct {
	ArtifactName ResourceName
	Timestamp    time.Time
}

func (ar ArtifactResource) UpdateTimestamp() time.Time {
	return ar.Timestamp
}

func (ar ArtifactResource) ResourceName() ResourceName {
	return ar.ArtifactName
}

func ListResources(ctx context.Context, client connection.Client, pattern, filter string) ([]ResourceInstance, error) {
	var result []ResourceInstance
	var err2 error

	// First try to match collection names.
	if api, err := names.ParseApiCollection(pattern); err == nil {
		err2 = core.ListAPIs(ctx, client, api, filter, generateApiHandler(&result))
	} else if version, err := names.ParseVersionCollection(pattern); err == nil {
		err2 = core.ListVersions(ctx, client, version, filter, generateVersionHandler(&result))
	} else if spec, err := names.ParseSpecCollection(pattern); err == nil {
		err2 = core.ListSpecs(ctx, client, spec, filter, generateSpecHandler(&result))
	} else if artifact, err := names.ParseArtifactCollection(pattern); err == nil {
		err2 = core.ListArtifacts(ctx, client, artifact, filter, false, generateArtifactHandler(&result))
	}

	// Then try to match resource names.
	if api, err := names.ParseApi(pattern); err == nil {
		err2 = core.ListAPIs(ctx, client, api, filter, generateApiHandler(&result))
	} else if version, err := names.ParseVersion(pattern); err == nil {
		err2 = core.ListVersions(ctx, client, version, filter, generateVersionHandler(&result))
	} else if spec, err := names.ParseSpec(pattern); err == nil {
		err2 = core.ListSpecs(ctx, client, spec, filter, generateSpecHandler(&result))
	} else if artifact, err := names.ParseArtifact(pattern); err == nil {
		err2 = core.ListArtifacts(ctx, client, artifact, filter, false, generateArtifactHandler(&result))
	}

	if err2 != nil {
		return nil, err2
	}

	return result, nil
}

func generateApiHandler(result *[]ResourceInstance) func(*rpc.Api) error {
	return func(api *rpc.Api) error {
		name, err := names.ParseApi(api.GetName())
		if err != nil {
			return err
		}

		(*result) = append((*result), ApiResource{
			ApiName:   ApiName{Name: name},
			Timestamp: api.UpdateTime.AsTime(),
		})

		return nil
	}
}

func generateVersionHandler(result *[]ResourceInstance) func(*rpc.ApiVersion) error {
	return func(version *rpc.ApiVersion) error {
		name, err := names.ParseVersion(version.GetName())
		if err != nil {
			return err
		}

		(*result) = append((*result), VersionResource{
			VersionName: VersionName{Name: name},
			Timestamp:   version.UpdateTime.AsTime(),
		})

		return nil
	}
}

func generateSpecHandler(result *[]ResourceInstance) func(*rpc.ApiSpec) error {
	return func(spec *rpc.ApiSpec) error {
		name, err := names.ParseSpec(spec.GetName())
		if err != nil {
			return err
		}

		(*result) = append((*result), SpecResource{
			SpecName:  SpecName{Name: name},
			Timestamp: spec.RevisionUpdateTime.AsTime(),
		})

		return nil
	}
}

func generateArtifactHandler(result *[]ResourceInstance) func(*rpc.Artifact) error {
	return func(artifact *rpc.Artifact) error {
		name, err := names.ParseArtifact(artifact.GetName())
		if err != nil {
			return err
		}

		(*result) = append((*result), ArtifactResource{
			ArtifactName: ArtifactName{Name: name},
			Timestamp:    artifact.UpdateTime.AsTime(),
		})

		return nil
	}
}
