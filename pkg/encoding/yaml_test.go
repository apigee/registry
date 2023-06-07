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

package encoding

import (
	"testing"

	"github.com/apigee/registry/pkg/application/apihub"
	"gopkg.in/yaml.v3"
)

func TestYAML(t *testing.T) {
	m := &apihub.DisplaySettings{
		Description:     "Display settings",
		Organization:    "ACME APIs",
		ApiGuideEnabled: true,
		ApiScoreEnabled: false,
	}
	node, err := NodeForMessage(m)
	if err != nil {
		t.Fatalf("Failed to encode message: %s", err)
	}
	a := &Artifact{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "Settings",
			Metadata: Metadata{
				Name: "settings",
			},
		},
		Data: *node,
	}
	bytes, err := EncodeYAML(a)
	if err != nil {
		t.Fatalf("Failed to encode artifact: %s", err)
	}
	expectedYAML := `apiVersion: apigeeregistry/v1
kind: Settings
metadata:
  name: settings
data:
  description: Display settings
  organization: ACME APIs
  apiGuideEnabled: true
  apiScoreEnabled: false
`
	if string(bytes) != expectedYAML {
		t.Errorf("Unexpected encoding. Got %s, expected %s", string(bytes), expectedYAML)
	}
	err = yaml.Unmarshal(bytes, node)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %s", err)
	}
	StyleForJSON(node)
	bytes, err = EncodeYAML(node)
	if err != nil {
		t.Fatalf("Failed to re-encode artifact: %s", err)
	}
	expectedJSON := `{"apiVersion": "apigeeregistry/v1", "kind": "Settings", "metadata": {"name": "settings"}, "data": {"description": "Display settings", "organization": "ACME APIs", "apiGuideEnabled": true, "apiScoreEnabled": false}}
`
	if string(bytes) != expectedJSON {
		t.Errorf("Unexpected encoding. Got %s, expected %s", string(bytes), expectedJSON)
	}
}
