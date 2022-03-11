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
type ResourceName interface {
	GetArtifact() string
	GetSpec() string
	GetVersion() string
	GetApi() string
	String() string
	GetParent() string
}

type SpecName struct {
	Spec names.Spec
}

func (s SpecName) GetArtifact() string {
	return ""
}

func (s SpecName) GetSpec() string {
	return s.Spec.String()
}

func (s SpecName) GetVersion() string {
	return s.Spec.Version().String()
}

func (s SpecName) GetApi() string {
	return s.Spec.Api().String()
}

func (s SpecName) String() string {
	return s.Spec.String()
}

func (s SpecName) GetParent() string {
	return s.Spec.Parent()
}

type VersionName struct {
	Version names.Version
}

func (v VersionName) GetArtifact() string {
	return ""
}

func (v VersionName) GetSpec() string {
	return ""
}

func (v VersionName) GetVersion() string {
	return v.Version.String()
}

func (v VersionName) GetApi() string {
	return v.Version.Api().String()
}

func (v VersionName) String() string {
	return v.Version.String()
}

func (v VersionName) GetParent() string {
	return v.Version.Parent()
}

type ApiName struct {
	Api names.Api
}

func (a ApiName) GetArtifact() string {
	return ""
}

func (a ApiName) GetSpec() string {
	return ""
}

func (a ApiName) GetVersion() string {
	return ""
}

func (a ApiName) GetApi() string {
	return a.Api.String()
}

func (a ApiName) String() string {
	return a.Api.String()
}

func (a ApiName) GetParent() string {
	return a.Api.Parent()
}

type ArtifactName struct {
	Artifact names.Artifact
}

func (ar ArtifactName) GetArtifact() string {
	return ar.Artifact.String()
}

func (ar ArtifactName) GetSpec() string {
	specPattern := names.Spec{
		ProjectID: ar.Artifact.ProjectID(),
		ApiID:     ar.Artifact.ApiID(),
		VersionID: ar.Artifact.VersionID(),
		SpecID:    ar.Artifact.SpecID(),
	}

	// Validate the generated name
	if spec, err := names.ParseSpec(specPattern.String()); err == nil {
		return spec.String()
	} else if _, err := names.ParseSpecCollection(specPattern.String()); err == nil {
		return spec.String()
	}

	return ""
}

func (ar ArtifactName) GetVersion() string {
	versionPattern := names.Version{
		ProjectID: ar.Artifact.ProjectID(),
		ApiID:     ar.Artifact.ApiID(),
		VersionID: ar.Artifact.VersionID(),
	}
	// Validate the generated name
	if version, err := names.ParseVersion(versionPattern.String()); err == nil {
		return version.String()
	} else if _, err := names.ParseVersionCollection(versionPattern.String()); err == nil {
		return version.String()
	}

	return ""
}

func (ar ArtifactName) GetApi() string {
	apiPattern := names.Api{
		ProjectID: ar.Artifact.ProjectID(),
		ApiID:     ar.Artifact.ApiID(),
	}
	// Validate the generated name
	if _, err := names.ParseApi(apiPattern.String()); err == nil {
		return apiPattern.String()
	} else if _, err := names.ParseApiCollection(apiPattern.String()); err == nil {
		return apiPattern.String()
	}

	return ""
}

func (ar ArtifactName) String() string {
	return ar.Artifact.String()
}

func (ar ArtifactName) GetParent() string {
	return ar.Artifact.Parent()
}
