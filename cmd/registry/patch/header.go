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

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const RegistryV1 = "apigeeregistry/v1"

type Header struct {
	ApiVersion string   `yaml:"apiVersion,omitempty"`
	Kind       string   `yaml:"kind,omitempty"`
	Metadata   Metadata `yaml:"metadata"`
}

type Metadata struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

func readHeader(info *yaml.Node) (*Header, error) {
	header := &Header{}
	var err error

	for _, node := range info.Content {
		if node.Kind == yaml.MappingNode {
			for i := 0; i < len(node.Content); i += 2 {
				key := node.Content[i]
				value := node.Content[i+1]
				switch key.Value {
				case "apiVersion":
					header.ApiVersion = value.Value
					if header.ApiVersion != RegistryV1 {
						return header, fmt.Errorf("unsupported API version: %s", header.ApiVersion)
					}
				case "kind":
					header.Kind = value.Value
				case "metadata":
					for i := 0; i < len(value.Content); i += 2 {
						key := node.Content[i]
						value := node.Content[i+1]
						switch key.Value {
						case "name":
							header.Metadata.Name = value.Value
						case "labels":
							header.Metadata.Labels, err = readDictionary(value)
							if err != nil {
								return nil, err
							}
						case "annotations":
							header.Metadata.Annotations, err = readDictionary(value)
							if err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
	}
	return header, nil
}

func readDictionary(node *yaml.Node) (map[string]string, error) {
	if node.Kind == yaml.MappingNode {
		m := make(map[string]string)
		for i := 0; i < len(node.Content); i += 2 {
			k := node.Content[i]
			if k.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("invalid map key (%s)", k.Value)
			}
			v := node.Content[i+1]
			if v.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("invalid map value (%s)", v.Value)
			}
			m[k.Value] = v.Value
		}
		return m, nil
	} else {
		return nil, fmt.Errorf("invalid dictionary %s", node.Value)
	}
}
