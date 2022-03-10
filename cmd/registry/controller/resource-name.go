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
	getArtifact() string
	getSpec() string
	getVersion() string
	getApi() string
	string() string
	getParent() string
}

type specName struct {
	spec names.Spec
}

func (s specName) getArtifact() string {
	return ""
}

func (s specName) getSpec() string {
	return s.spec.String()
}

func (s specName) getVersion() string {
	return s.spec.Version().String()
}

func (s specName) getApi() string {
	return s.spec.Api().String()
}

func (s specName) string() string {
	return s.spec.String()
}

func (s specName) getParent() string {
	return s.spec.Parent()
}

type versionName struct {
	version names.Version
}

func (v versionName) getArtifact() string {
	return ""
}

func (v versionName) getSpec() string {
	return ""
}

func (v versionName) getVersion() string {
	return v.version.String()
}

func (v versionName) getApi() string {
	return v.version.Api().String()
}

func (v versionName) string() string {
	return v.version.String()
}

func (v versionName) getParent() string {
	return v.version.Parent()
}

type apiName struct {
	api names.Api
}

func (a apiName) getArtifact() string {
	return ""
}

func (a apiName) getSpec() string {
	return ""
}

func (a apiName) getVersion() string {
	return ""
}

func (a apiName) getApi() string {
	return a.api.String()
}

func (a apiName) string() string {
	return a.api.String()
}

func (a apiName) getParent() string {
	return a.api.Parent()
}

type artifactName struct {
	artifact names.Artifact
}

func (ar artifactName) getArtifact() string {
	return ar.artifact.String()
}

func (ar artifactName) getSpec() string {
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

func (ar artifactName) getVersion() string {
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

func (ar artifactName) getApi() string {
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

func (ar artifactName) string() string {
	return ar.artifact.String()
}

func (ar artifactName) getParent() string {
	return ar.artifact.Parent()
}
