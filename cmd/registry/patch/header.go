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

package patch

import (
	"github.com/apigee/registry/pkg/encoding"
	"gopkg.in/yaml.v3"
)

const RegistryV1 = "apigeeregistry/v1"

func readHeader(bytes []byte) (encoding.Header, error) {
	var header encoding.Header
	err := yaml.Unmarshal(bytes, &header)
	return header, err
}

func readHeaderWithItems(bytes []byte) (encoding.Header, yaml.Node, error) {
	type headerWithItems struct {
		encoding.Header `yaml:",inline"`
		Items           yaml.Node `yaml:"items,omitempty"`
	}
	var wrapper headerWithItems
	err := yaml.Unmarshal(bytes, &wrapper)
	return wrapper.Header, wrapper.Items, err
}
