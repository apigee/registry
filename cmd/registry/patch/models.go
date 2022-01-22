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

type Header struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

type Project struct {
	Header `yaml:",inline"`
	Spec   struct {
		DisplayName string     `yaml:"displayName"`
		Description string     `yaml:"description"`
		APIs        []API      `yaml:"apis"`
		Artifacts   []Artifact `yaml:"artifacts"`
	} `yaml:"spec"`
}

type API struct {
	Header `yaml:",inline"`
	Spec   struct {
		DisplayName    string          `yaml:"displayName"`
		Description    string          `yaml:"description"`
		APIVersions    []APIVersion    `yaml:"versions"`
		APIDeployments []APIDeployment `yaml:"deployments"`
	} `yaml:"spec"`
}

type APIVersion struct {
	Header `yaml:",inline"`
	Spec   struct {
		DisplayName string    `yaml:"displayName"`
		Description string    `yaml:"description"`
		APISpecs    []APISpec `yaml:"specs"`
	} `yaml:"spec"`
}

type APISpec struct {
	Header `yaml:",inline"`
	Spec   struct {
		FileName    string `yaml:"fileName"`
		Description string `yaml:"description"`
	} `yaml:"spec"`
}

type APIDeployment struct {
	Header `yaml:",inline"`
	Spec   struct {
		DisplayName string `yaml:"displayName"`
		Description string `yaml:"description"`
	} `yaml:"spec"`
}

type Artifact struct {
	Header `yaml:",inline"`
	Spec   struct {
		MimeType string `yaml:"mimeType"`
	} `yaml:"spec"`
}
