package core

import (
	discovery "github.com/googleapis/gnostic/discovery"
	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
)

func SummarizeOpenAPIv2Document(document *openapi_v2.Document) *metrics.Complexity {
	summary := &metrics.Complexity{}
	if document.Definitions != nil && document.Definitions.AdditionalProperties != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			summarizeOpenAPIv2Schema(summary, pair.Value)
		}
	}
	for _, pair := range document.Paths.Path {
		summary.PathCount++
		v := pair.Value
		if v.Get != nil {
			summary.GetCount++
		}
		if v.Post != nil {
			summary.PostCount++
		}
		if v.Put != nil {
			summary.PutCount++
		}
		if v.Delete != nil {
			summary.DeleteCount++
		}
	}
	return summary
}

func summarizeOpenAPIv2Schema(summary *metrics.Complexity, schema *openapi_v2.Schema) {
	summary.SchemaCount++
	if schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			summary.SchemaPropertyCount++
			summarizeOpenAPIv2Schema(summary, pair.Value)
		}
	}
}

func SummarizeOpenAPIv3Document(document *openapi_v3.Document) *metrics.Complexity {
	summary := &metrics.Complexity{}
	if document.Components != nil && document.Components.Schemas != nil {
		for _, pair := range document.Components.Schemas.AdditionalProperties {
			summarizeOpenAPIv3Schema(summary, pair.Value)
		}
	}
	for _, pair := range document.Paths.Path {
		summary.PathCount++
		v := pair.Value
		if v.Get != nil {
			summary.GetCount++
		}
		if v.Post != nil {
			summary.PostCount++
		}
		if v.Put != nil {
			summary.PutCount++
		}
		if v.Delete != nil {
			summary.DeleteCount++
		}
	}
	return summary
}

func summarizeOpenAPIv3Schema(summary *metrics.Complexity, schemaOrReference *openapi_v3.SchemaOrReference) {
	summary.SchemaCount++
	schema := schemaOrReference.GetSchema()
	if schema != nil && schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			summary.SchemaPropertyCount++
			summarizeOpenAPIv3Schema(summary, pair.Value)
		}
	}
}

func SummarizeDiscoveryDocument(document *discovery.Document) *metrics.Complexity {
	summary := &metrics.Complexity{}
	if document.Schemas != nil && document.Schemas.AdditionalProperties != nil {
		for _, pair := range document.Schemas.AdditionalProperties {
			summarizeDiscoverySchema(summary, pair.Value)
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
