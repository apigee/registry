// Copyright 2021 Google LLC
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

package controller

import (
	"github.com/apigee/registry/server/registry/names"
)

// This interface is used to describe resource patterns
// Example:
// projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml
// projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-
type resourceName interface {
	Artifact() string
	Spec() string
	Version() string
	Api() string
	String() string
	ParentName() resourceName
}

type specName struct {
	spec names.Spec
}

func (s specName) Artifact() string {
	return ""
}

func (s specName) Spec() string {
	return s.spec.String()
}

func (s specName) Version() string {
	return s.spec.Version().String()
}

func (s specName) Api() string {
	return s.spec.Api().String()
}

func (s specName) String() string {
	return s.spec.String()
}

func (s specName) ParentName() resourceName {
	// Validate the parent name aand return
	if version, err := names.ParseVersion(s.spec.Parent()); err == nil {
		return versionName{
			version: version,
		}
	} else if version, err := names.ParseVersionCollection(s.spec.Parent()); err == nil {
		return versionName{
			version: version,
		}
	}
	return versionName{}
}

type versionName struct {
	version names.Version
}

func (v versionName) Artifact() string {
	return ""
}

func (v versionName) Spec() string {
	return ""
}

func (v versionName) Version() string {
	return v.version.String()
}

func (v versionName) Api() string {
	return v.version.Api().String()
}

func (v versionName) String() string {
	return v.version.String()
}

func (v versionName) ParentName() resourceName {
	// Validate the parent name aand return
	if api, err := names.ParseApi(v.version.Parent()); err == nil {
		return apiName{
			api: api,
		}
	} else if api, err := names.ParseApiCollection(v.version.Parent()); err == nil {
		return apiName{
			api: api,
		}
	}

	return apiName{}
}

type apiName struct {
	api names.Api
}

func (a apiName) Artifact() string {
	return ""
}

func (a apiName) Spec() string {
	return ""
}

func (a apiName) Version() string {
	return ""
}

func (a apiName) Api() string {
	return a.api.String()
}

func (a apiName) String() string {
	return a.api.String()
}

func (a apiName) ParentName() resourceName {
	// Validate the parent name aand return
	if project, err := names.ParseProject(a.api.Parent()); err == nil {
		return projectName{
			project: project,
		}
	} else if project, err := names.ParseProjectWithLocation(a.api.Parent()); err == nil {
		return projectName{
			project: project,
		}
	} else if project, err := names.ParseProjectCollection(a.api.Parent()); err == nil {
		return projectName{
			project: project,
		}
	}

	return projectName{}
}

type projectName struct {
	project names.Project
}

func (p projectName) Artifact() string {
	return ""
}

func (p projectName) Spec() string {
	return ""
}

func (p projectName) Version() string {
	return ""
}

func (p projectName) Api() string {
	return ""
}

func (p projectName) String() string {
	return p.project.String()
}

func (p projectName) ParentName() resourceName {
	return nil
}

type artifactName struct {
	artifact names.Artifact
}

func (ar artifactName) Artifact() string {
	return ar.artifact.String()
}

func (ar artifactName) Spec() string {
	specPattern := names.Spec{
		ProjectID: ar.artifact.ProjectID(),
		ApiID:     ar.artifact.ApiID(),
		VersionID: ar.artifact.VersionID(),
		SpecID:    ar.artifact.SpecID(),
	}

	// Validate the generated name
	if spec, err := names.ParseSpec(specPattern.String()); err == nil {
		return spec.String()
	} else if _, err := names.ParseSpecCollection(specPattern.String()); err == nil {
		return spec.String()
	}

	return ""
}

func (ar artifactName) Version() string {
	versionPattern := names.Version{
		ProjectID: ar.artifact.ProjectID(),
		ApiID:     ar.artifact.ApiID(),
		VersionID: ar.artifact.VersionID(),
	}
	// Validate the generated name
	if version, err := names.ParseVersion(versionPattern.String()); err == nil {
		return version.String()
	} else if _, err := names.ParseVersionCollection(versionPattern.String()); err == nil {
		return version.String()
	}

	return ""
}

func (ar artifactName) Api() string {
	apiPattern := names.Api{
		ProjectID: ar.artifact.ProjectID(),
		ApiID:     ar.artifact.ApiID(),
	}
	// Validate the generated name
	if _, err := names.ParseApi(apiPattern.String()); err == nil {
		return apiPattern.String()
	} else if _, err := names.ParseApiCollection(apiPattern.String()); err == nil {
		return apiPattern.String()
	}

	return ""
}

func (ar artifactName) String() string {
	return ar.artifact.String()
}

func (ar artifactName) ParentName() resourceName {
	parent := ar.artifact.Parent()
	// First try to match collection names.
	if project, err := names.ParseProjectCollection(parent); err == nil {
		return projectName{
			project: project,
		}
	} else if api, err := names.ParseApiCollection(parent); err == nil {
		return apiName{
			api: api,
		}
	} else if version, err := names.ParseVersionCollection(parent); err == nil {
		return versionName{
			version: version,
		}
	} else if spec, err := names.ParseSpecCollection(parent); err == nil {
		return specName{
			spec: spec,
		}
	}

	// Then try to match resource names.
	if project, err := names.ParseProject(parent); err == nil {
		return projectName{
			project: project,
		}
	} else if project, err := names.ParseProjectWithLocation(parent); err == nil {
		return projectName{
			project: project,
		}
	} else if api, err := names.ParseApi(parent); err == nil {
		return apiName{
			api: api,
		}
	} else if version, err := names.ParseVersion(parent); err == nil {
		return versionName{
			version: version,
		}
	} else if spec, err := names.ParseSpec(parent); err == nil {
		return specName{
			spec: spec,
		}
	}

	return nil
}
