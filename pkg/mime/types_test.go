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

package mime

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIMimeTypes(t *testing.T) {
	tests := []struct {
		name        string
		compression string
		version     string
	}{
		{
			compression: "",
			version:     "2",
			name:        "application/x.openapi;version=2",
		},
		{
			compression: "",
			version:     "3",
			name:        "application/x.openapi;version=3",
		},
		{
			compression: "+gzip",
			version:     "2",
			name:        "application/x.openapi+gzip;version=2",
		},
		{
			compression: "+gzip",
			version:     "3",
			name:        "application/x.openapi+gzip;version=3",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := OpenAPIMimeType(test.compression, test.version)
			if value != test.name {
				t.Errorf("expected mime type %s got %s", test.name, value)
			}
			if strings.HasPrefix(test.version, "2") != IsOpenAPIv2(value) {
				t.Errorf("%s OpenAPI version is incorrectly recognized", value)
			}
			if strings.HasPrefix(test.version, "3") != IsOpenAPIv3(value) {
				t.Errorf("%s OpenAPI version is incorrectly recognized", value)
			}
			if IsDiscovery(value) {
				t.Errorf("%s is incorrectly identified as a discovery type", value)
			}
			if IsProto(value) {
				t.Errorf("%s is incorrectly identified as a protobuf type", value)
			}
			if IsGZipCompressed(value) != (test.compression == "+gzip") {
				t.Errorf("%s compression is incorrectly recognized", value)
			}
			if IsZipArchive(value) {
				t.Errorf("%s is incorrectly recognized as a zip archive", value)
			}
			if IsGZipCompressed(value) {
				unzipped := GUnzippedType(value)
				if IsGZipCompressed(unzipped) {
					t.Errorf("failed to remove compression from type %q", value)
				}
			}
		})
	}
}

func TestDiscoveryMimeTypes(t *testing.T) {
	tests := []struct {
		name        string
		compression string
		version     string
	}{
		{
			compression: "",
			name:        "application/x.discovery",
		},
		{
			compression: "+gzip",
			name:        "application/x.discovery+gzip",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := DiscoveryMimeType(test.compression)
			if value != test.name {
				t.Errorf("expected mime type %s got %s", test.name, value)
			}
			if IsOpenAPIv2(value) {
				t.Errorf("%s is incorrectly recognized as OpenAPI v2", value)
			}
			if IsOpenAPIv3(value) {
				t.Errorf("%s is incorrectly recognized as OpenAPI v3", value)
			}
			if !IsDiscovery(value) {
				t.Errorf("%s is not recognized as a discovery type", value)
			}
			if IsProto(value) {
				t.Errorf("%s is incorrectly identified as a protobuf type", value)
			}
			if IsGZipCompressed(value) != (test.compression == "+gzip") {
				t.Errorf("%s compression is incorrectly recognized", value)
			}
			if IsZipArchive(value) {
				t.Errorf("%s is incorrectly recognized as a zip archive", value)
			}
		})
	}
}

func TestProtobufMimeTypes(t *testing.T) {
	tests := []struct {
		name        string
		compression string
		version     string
	}{
		{
			compression: "",
			name:        "application/x.protobuf",
		},
		{
			compression: "+zip",
			name:        "application/x.protobuf+zip",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := ProtobufMimeType(test.compression)
			if value != test.name {
				t.Errorf("expected mime type %s got %s", test.name, value)
			}
			if IsOpenAPIv2(value) {
				t.Errorf("%s is incorrectly recognized as OpenAPI v2", value)
			}
			if IsOpenAPIv3(value) {
				t.Errorf("%s is incorrectly recognized as OpenAPI v3", value)
			}
			if IsDiscovery(value) {
				t.Errorf("%s is incorrectly recognized as a discovery type", value)
			}
			if !IsProto(value) {
				t.Errorf("%s is not recognized as a protobuf type", value)
			}
			if IsGZipCompressed(value) != (test.compression == "+gzip") {
				t.Errorf("%s compression is incorrectly recognized", value)
			}
			if IsZipArchive(value) != (test.compression == "+zip") {
				t.Errorf("%s compression is incorrectly recognized", value)
			}
		})
	}
}

func TestProtobufMessageTypes(t *testing.T) {
	tests := []struct {
		kind        string
		messageType string
		mimeType    string
	}{
		{
			kind:        "ApiSpecExtensionList",
			messageType: "google.cloud.apigeeregistry.v1.apihub.ApiSpecExtensionList",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.ApiSpecExtensionList",
		},
		{
			kind:        "DisplaySettings",
			messageType: "google.cloud.apigeeregistry.v1.apihub.DisplaySettings",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.DisplaySettings",
		},
		{
			kind:        "Lifecycle",
			messageType: "google.cloud.apigeeregistry.v1.apihub.Lifecycle",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.Lifecycle",
		},
		{
			kind:        "ReferenceList",
			messageType: "google.cloud.apigeeregistry.v1.apihub.ReferenceList",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.ReferenceList",
		},
		{
			kind:        "TaxonomyList",
			messageType: "google.cloud.apigeeregistry.v1.apihub.TaxonomyList",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.TaxonomyList",
		},
		{
			kind:        "Manifest",
			messageType: "google.cloud.apigeeregistry.v1.controller.Manifest",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.controller.Manifest",
		},
		{
			kind:        "Receipt",
			messageType: "google.cloud.apigeeregistry.v1.controller.Receipt",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.controller.Receipt",
		},
		{
			kind:        "Score",
			messageType: "google.cloud.apigeeregistry.v1.scoring.Score",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.Score",
		},
		{
			kind:        "ScoreDefinition",
			messageType: "google.cloud.apigeeregistry.v1.scoring.ScoreDefinition",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreDefinition",
		},
		{
			kind:        "ScoreCard",
			messageType: "google.cloud.apigeeregistry.v1.scoring.ScoreCard",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreCard",
		},
		{
			kind:        "ScoreCardDefinition",
			messageType: "google.cloud.apigeeregistry.v1.scoring.ScoreCardDefinition",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreCardDefinition",
		},
		{
			kind:        "StyleGuide",
			messageType: "google.cloud.apigeeregistry.v1.style.StyleGuide",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.StyleGuide",
		},
		{
			kind:        "ConformanceReport",
			messageType: "google.cloud.apigeeregistry.v1.style.ConformanceReport",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.ConformanceReport",
		},
		{
			kind:        "Lint",
			messageType: "google.cloud.apigeeregistry.v1.style.Lint",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
		},
		{
			kind:        "Complexity",
			messageType: "gnostic.metrics.Complexity",
			mimeType:    "application/octet-stream;type=gnostic.metrics.Complexity",
		},
		{
			kind:        "Vocabulary",
			messageType: "gnostic.metrics.Vocabulary",
			mimeType:    "application/octet-stream;type=gnostic.metrics.Vocabulary",
		},
		{
			kind:        "FieldSet",
			messageType: "google.cloud.apigeeregistry.v1.apihub.FieldSet",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.FieldSet",
		},
		{
			kind:        "FieldSetDefinition",
			messageType: "google.cloud.apigeeregistry.v1.apihub.FieldSetDefinition",
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.FieldSetDefinition",
		},
	}
	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			var err error
			value := MimeTypeForMessageType(test.messageType)
			if value != test.mimeType {
				t.Errorf("incorrect mime type for message type, expected %s got %s", test.mimeType, value)
			}
			value = MimeTypeForKind(test.kind)
			if value != test.mimeType {
				t.Errorf("incorrect mime type for kind, expected %s got %s", test.mimeType, value)
			}
			value, err = MessageTypeForMimeType(test.mimeType)
			if err != nil {
				t.Errorf("Error getting message type for mime type %s", err)
			}
			if value != test.messageType {
				t.Errorf("incorrect message type, expected %s got %s", test.messageType, value)
			}
			value = KindForMimeType(test.mimeType)
			if value != test.kind {
				t.Errorf("incorrect kind, expected %s got %s", test.kind, value)
			}
			message1, err := MessageForMimeType(test.mimeType)
			if err != nil {
				t.Errorf("Error getting message for mime type %s", err)
			}
			type1 := filepath.Ext(fmt.Sprintf("%T", message1))[1:]
			if type1 != test.kind {
				t.Errorf("incorrect message for mime type, expected %s got %s", test.kind, type1)
			}
			message2, err := MessageForKind(test.kind)
			if err != nil {
				t.Errorf("Error getting message for kind %s", err)
			}
			type2 := filepath.Ext(fmt.Sprintf("%T", message2))[1:]
			if type2 != test.kind {
				t.Errorf("incorrect message for kind, expected %s got %s", test.kind, type2)
			}

			// yaml variations of the above
			yamlType := strings.Replace(test.mimeType, "octet-stream", "yaml", 1)
			value = YamlMimeTypeForKind(test.kind)
			if value != yamlType {
				t.Errorf("incorrect mime type for kind, expected %s got %s", yamlType, value)
			}
			value = KindForMimeType(yamlType)
			if value != test.kind {
				t.Errorf("incorrect kind, expected %s got %s", test.kind, value)
			}
			value, err = MessageTypeForMimeType(yamlType)
			if err != nil {
				t.Errorf("Error getting message type for mime type %s", err)
			}
			if value != test.messageType {
				t.Errorf("incorrect message type, expected %s got %s", test.messageType, value)
			}
		})
	}
}

func TestYamlMessageTypes(t *testing.T) {
	tests := []struct {
		kind     string
		mimeType string
	}{
		{
			kind:     "Sample",
			mimeType: "application/yaml;type=Sample",
		},
		{
			kind:     "",
			mimeType: "application/yaml",
		},
	}
	for _, test := range tests {
		t.Run(test.mimeType, func(t *testing.T) {
			value := MimeTypeForKind(test.kind)
			if value != test.mimeType {
				t.Errorf("incorrect mime type for kind, expected %s got %s", test.mimeType, value)
			}
			value = YamlMimeTypeForKind(test.kind)
			if value != test.mimeType {
				t.Errorf("incorrect mime type for kind, expected %s got %s", test.mimeType, value)
			}
			value = KindForMimeType(test.mimeType)
			if value != test.kind {
				t.Errorf("incorrect kind, expected %s got %s", test.kind, value)
			}
		})
	}
}

func TestInvalidTypes(t *testing.T) {
	tests := []struct {
		kind              string
		mimeType          string
		kindIsInvalid     bool
		mimeTypeIsInvalid bool
	}{
		{
			kind:              "ScoreCard",
			mimeType:          "application/binary;type=google.cloud.apigeeregistry.v1.scoring.ScoreCard",
			kindIsInvalid:     false,
			mimeTypeIsInvalid: true,
		},
		{
			kind:              "Invalid",
			mimeType:          "application/octet-stream;type=google.cloud.apigeeregistry.v1.invalid.Invalid",
			kindIsInvalid:     true,
			mimeTypeIsInvalid: false,
		},
	}
	for _, test := range tests {
		t.Run(test.mimeType, func(t *testing.T) {
			var err error
			_, err = MessageTypeForMimeType(test.mimeType)
			if err == nil && test.mimeTypeIsInvalid {
				t.Errorf("Did not obtain expected error getting message type for invalid mime type %s", test.mimeType)
			}
			_, err = MessageForMimeType(test.mimeType)
			if err == nil && test.mimeTypeIsInvalid {
				t.Errorf("Did not obtain expected error getting message for invalid mime type %s", test.mimeType)
			}
			_, err = MessageForKind(test.kind)
			if err == nil && test.kindIsInvalid {
				t.Errorf("Did not obtain expected error getting message for invalid kind %s", test.kind)
			}
		})
	}
}

func TestPrintableTypes(t *testing.T) {
	tests := []struct {
		mimeType    string
		isPrintable bool
	}{
		{
			mimeType:    "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.StyleGuide",
			isPrintable: false,
		},
		{
			mimeType:    "text/plain",
			isPrintable: true,
		},
		{
			mimeType:    "application/json;type=google.cloud.apigeeregistry.v1.style.StyleGuide",
			isPrintable: true,
		},
		{
			mimeType:    "application/yaml;type=Struct",
			isPrintable: true,
		},
	}
	for _, test := range tests {
		t.Run(test.mimeType, func(t *testing.T) {
			value := IsPrintableType(test.mimeType)
			if value != test.isPrintable {
				t.Errorf("Did not obtain expected value for isPrintable: expected %t got %t", test.isPrintable, value)
			}
		})
	}
}
