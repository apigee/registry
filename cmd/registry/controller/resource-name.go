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
	GetName() string
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

func (s SpecName) GetName() string {
	return s.Spec.String()
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

func (v VersionName) GetName() string {
	return v.Version.String()
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

func (a ApiName) GetName() string {
	return a.Api.String()
}

type ArtifactName struct {
	Artifact names.Artifact
}

func (ar ArtifactName) GetArtifact() string {
	return ar.Artifact.String()
}

func (ar ArtifactName) GetSpec() string {
	specName := names.Spec{
		ProjectID: ar.Artifact.ProjectID(),
		ApiID:     ar.Artifact.ApiID(),
		VersionID: ar.Artifact.VersionID(),
		SpecID:    ar.Artifact.SpecID(),
	}
	// if err := specName.Validate(); err != nil {
	// 	return ""
	// }

	return specName.String()
}

func (ar ArtifactName) GetVersion() string {
	versionName := names.Version{
		ProjectID: ar.Artifact.ProjectID(),
		ApiID:     ar.Artifact.ApiID(),
		VersionID: ar.Artifact.VersionID(),
	}
	// if err := versionName.Validate(); err != nil {
	// 	return ""
	// }

	return versionName.String()
}

func (ar ArtifactName) GetApi() string {
	apiName := names.Api{
		ProjectID: ar.Artifact.ProjectID(),
		ApiID:     ar.Artifact.ApiID(),
	}
	// if err := apiName.Validate(); err != nil {
	// 	return ""
	// }

	return apiName.String()
}

func (ar ArtifactName) GetName() string {
	return ar.Artifact.String()
}
