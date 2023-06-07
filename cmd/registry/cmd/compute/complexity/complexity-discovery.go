// Copyright 2020 Google LLC.
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

package complexity

import (
	discovery "github.com/google/gnostic/discovery"
	metrics "github.com/google/gnostic/metrics"
)

func SummarizeDiscoveryDocument(document *discovery.Document) *metrics.Complexity {
	summary := &metrics.Complexity{}
	if document.Schemas != nil && document.Schemas.AdditionalProperties != nil {
		for _, pair := range document.Schemas.AdditionalProperties {
			summarizeDiscoverySchema(summary, pair.Value)
		}
	}
	if document.Resources != nil {
		for _, pair := range document.Resources.AdditionalProperties {
			summarizeDiscoveryResource(summary, pair.Value)
		}
	}
	if document.Methods != nil {
		for _, pair := range document.Methods.AdditionalProperties {
			summary.PathCount++
			v := pair.Value
			switch v.HttpMethod {
			case "GET":
				summary.GetCount++
			case "POST":
				summary.PostCount++
			case "PUT":
				summary.PutCount++
			case "DELETE":
				summary.DeleteCount++
			}
		}
	}
	return summary
}

func summarizeDiscoverySchema(summary *metrics.Complexity, schema *discovery.Schema) {
	summary.SchemaCount++
	if schema != nil && schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			summary.SchemaPropertyCount++
			summarizeDiscoverySchema(summary, pair.Value)
		}
	}
}

func summarizeDiscoveryResource(summary *metrics.Complexity, resource *discovery.Resource) {
	if resource.Resources != nil {
		for _, pair := range resource.Resources.AdditionalProperties {
			summarizeDiscoveryResource(summary, pair.Value)
		}
	}
	if resource.Methods != nil {
		for _, pair := range resource.Methods.AdditionalProperties {
			summary.PathCount++
			v := pair.Value
			switch v.HttpMethod {
			case "GET":
				summary.GetCount++
			case "POST":
				summary.PostCount++
			case "PUT":
				summary.PutCount++
			case "DELETE":
				summary.DeleteCount++
			}
		}
	}
}
