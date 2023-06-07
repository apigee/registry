// Copyright 2022 Google LLC.
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

package encoding

type ApiDeployment struct {
	Header `yaml:",inline"`
	Data   ApiDeploymentData `yaml:"data"`
}

type ApiDeploymentData struct {
	DisplayName        string      `yaml:"displayName,omitempty"`
	Description        string      `yaml:"description,omitempty"`
	ApiSpecRevision    string      `yaml:"apiSpecRevision,omitempty"`
	EndpointURI        string      `yaml:"endpointURI,omitempty"`
	ExternalChannelURI string      `yaml:"externalChannelURI,omitempty"`
	IntendedAudience   string      `yaml:"intendedAudience,omitempty"`
	AccessGuidance     string      `yaml:"accessGuidance,omitempty"`
	Artifacts          []*Artifact `yaml:"artifacts,omitempty"`
}
