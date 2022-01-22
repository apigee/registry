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

package patch

const REGISTRY_V1 = "registry/v1"

type Header struct {
	APIVersion string   `yaml:"apiVersion,omitempty"`
	Kind       string   `yaml:"kind,omitempty"`
	Metadata   Metadata `yaml:"metadata"`
}

type Metadata struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type Project struct {
	Header `yaml:",inline"`
	Spec   struct {
		DisplayName string      `yaml:"displayName,omitempty"`
		Description string      `yaml:"description,omitempty"`
		APIs        []*API      `yaml:"apis,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"body"`
}

type API struct {
	Header `yaml:",inline"`
	Spec   struct {
		DisplayName           string           `yaml:"displayName,omitempty"`
		Description           string           `yaml:"description,omitempty"`
		Availability          string           `yaml:"availability,omitempty"`
		RecommendedVersion    string           `yaml:"recommendedVersion,omitempty"`
		RecommendedDeployment string           `yaml:"recommendedDeployment,omitempty"`
		APIVersions           []*APIVersion    `yaml:"versions,omitempty"`
		APIDeployments        []*APIDeployment `yaml:"deployments,omitempty"`
		Artifacts             []*Artifact      `yaml:"artifacts,omitempty"`
	} `yaml:"body"`
}

type APIVersion struct {
	Header `yaml:",inline"`
	Spec   struct {
		DisplayName string      `yaml:"displayName,omitempty"`
		Description string      `yaml:"description,omitempty"`
		APISpecs    []*APISpec  `yaml:"specs,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"body"`
}

type APISpec struct {
	Header `yaml:",inline"`
	Spec   struct {
		FileName    string      `yaml:"fileName,omitempty"`
		Description string      `yaml:"description,omitempty"`
		MimeType    string      `yaml:"mimeType,omitempty"`
		SourceURI   string      `yaml:"sourceURI,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"body"`
}

type APIDeployment struct {
	Header `yaml:",inline"`
	Spec   struct {
		DisplayName string      `yaml:"displayName,omitempty"`
		Description string      `yaml:"description,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"body"`
}

type Artifact struct {
	Header `yaml:",inline"`
	Spec   struct {
		MimeType string `yaml:"mimeType,omitempty"`
	} `yaml:"body"`
}
