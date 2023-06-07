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
	"regexp"
	"strings"

	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/application/controller"
	"github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/pkg/application/style"
	metrics "github.com/google/gnostic/metrics"
	"google.golang.org/protobuf/proto"
)

// OpenAPIMimeType returns a MIME type for an OpenAPI description of an API.
func OpenAPIMimeType(compression, version string) string {
	return fmt.Sprintf("application/x.openapi%s;version=%s", compression, version)
}

// DiscoveryMimeType returns a MIME type for a Discovery description of an API.
func DiscoveryMimeType(compression string) string {
	return fmt.Sprintf("application/x.discovery%s", compression)
}

// ProtobufMimeType returns a MIME type for a Protocol Buffers description of an API.
func ProtobufMimeType(compression string) string {
	return fmt.Sprintf("application/x.protobuf%s", compression)
}

// IsOpenAPIv2 returns true if a MIME type represents an OpenAPI v2 spec.
func IsOpenAPIv2(mimeType string) bool {
	return strings.Contains(mimeType, "openapi") &&
		strings.Contains(mimeType, "version=2")
}

// IsOpenAPIv3 returns true if a MIME type represents an OpenAPI v3 spec.
func IsOpenAPIv3(mimeType string) bool {
	return strings.Contains(mimeType, "openapi") &&
		strings.Contains(mimeType, "version=3")
}

// IsDiscovery returns true if a MIME type represents a Google API Discovery document.
func IsDiscovery(mimeType string) bool {
	return strings.Contains(mimeType, "discovery")
}

// IsProto returns true if a MIME type represents a Protocol Buffers Language API description.
func IsProto(mimeType string) bool {
	return strings.Contains(mimeType, "proto")
}

// IsGZipCompressed returns true if a MIME type represents a type compressed with GZip encoding.
func IsGZipCompressed(mimeType string) bool {
	return strings.Contains(mimeType, "+gzip")
}

// GUnzippedType returns an equivalent uncompressed MIME type for a compressed type.
func GUnzippedType(mimeType string) string {
	return strings.Replace(mimeType, "+gzip", "", 1)
}

// IsZipArchive returns true if a MIME type represents a type stored as a multifile Zip archive.
func IsZipArchive(mimeType string) bool {
	return strings.Contains(mimeType, "+zip")
}

// IsPrintableType returns true if the corresponding contents can be treated as a string.
func IsPrintableType(mimeType string) bool {
	return mimeType == "text/plain" ||
		strings.HasPrefix(mimeType, "application/yaml") ||
		strings.HasPrefix(mimeType, "application/json")
}

func IsYamlKind(mimeType string) bool {
	return strings.HasPrefix(mimeType, "application/yaml;type=") && KindForMimeType(mimeType) != ""
}

// MimeTypeForMessageType returns a MIME type that represents a Protocol Buffer message type.
func MimeTypeForMessageType(protoType string) string {
	return fmt.Sprintf("application/octet-stream;type=%s", protoType)
}

// MessageTypeForMimeType returns the Protocol Buffer message type represented by a MIME type.
func MessageTypeForMimeType(protoType string) (string, error) {
	re1 := regexp.MustCompile("^application/yaml;type=(.*)$")
	m1 := re1.FindStringSubmatch(protoType)
	if m1 != nil && len(m1[1]) > 0 {
		return strings.TrimSuffix(m1[1], "+gzip"), nil
	}

	re := regexp.MustCompile("^application/octet-stream;type=(.*)$")
	m := re.FindStringSubmatch(protoType)
	if m == nil || len(m) < 2 || len(m[1]) == 0 {
		return "", fmt.Errorf("invalid Protocol Buffer type: %s", protoType)
	}
	return strings.TrimSuffix(m[1], "+gzip"), nil
}

// KindForMimeType returns the name to be used as the "kind" of an exported artifact.
func KindForMimeType(mimeType string) string {
	if strings.HasPrefix(mimeType, "application/yaml;type=") {
		typeParameter := strings.TrimPrefix(mimeType, "application/yaml;type=")
		parts := strings.Split(typeParameter, ".")
		return parts[len(parts)-1]
	} else if strings.HasPrefix(mimeType, "application/octet-stream;type=") {
		typeParameter := strings.TrimPrefix(mimeType, "application/octet-stream;type=")
		parts := strings.Split(typeParameter, ".")
		return parts[len(parts)-1]
	} else {
		return ""
	}
}

// MessageForMimeType returns an instance of the message that represents the specified MIME type.
func MessageForMimeType(mimeType string) (proto.Message, error) {
	messageType, err := MessageTypeForMimeType(mimeType)
	if err != nil {
		return nil, err
	}
	f := artifactMessageTypes[messageType]
	if f == nil {
		return nil, fmt.Errorf("unsupported message type %s", messageType)
	}
	return f(), nil
}

// MessageForKind returns an instance of the message that represents the specified kind.
func MessageForKind(kind string) (proto.Message, error) {
	for k, v := range artifactMessageTypes {
		if strings.HasSuffix(k, "."+kind) {
			return v(), nil
		}
	}
	return nil, fmt.Errorf("unsupported kind %s", kind)
}

// MimeTypeForKind returns the mime type that corresponds to a kind.
func MimeTypeForKind(kind string) string {
	if kind == "" {
		return "application/yaml"
	}
	for k := range artifactMessageTypes {
		if strings.HasSuffix(k, "."+kind) {
			return fmt.Sprintf("application/octet-stream;type=%s", k)
		}
	}
	return fmt.Sprintf("application/yaml;type=%s", kind)
}

// YamlMimeTypeForKind returns a YAML mime type that corresponds to a kind.
func YamlMimeTypeForKind(kind string) string {
	if kind == "" {
		return "application/yaml"
	}
	for k := range artifactMessageTypes {
		if strings.HasSuffix(k, "."+kind) {
			return fmt.Sprintf("application/yaml;type=%s", k)
		}
	}
	return fmt.Sprintf("application/yaml;type=%s", kind)
}

// messageFactory represents functions that construct message structs.
type messageFactory func() proto.Message

// artifactMessageTypes is the single source of truth for protobuf types that the registry tool supports in artifact YAML files.
var artifactMessageTypes map[string]messageFactory = map[string]messageFactory{
	"google.cloud.apigeeregistry.v1.apihub.ApiSpecExtensionList": func() proto.Message { return new(apihub.ApiSpecExtensionList) },
	"google.cloud.apigeeregistry.v1.apihub.DisplaySettings":      func() proto.Message { return new(apihub.DisplaySettings) },
	"google.cloud.apigeeregistry.v1.apihub.Lifecycle":            func() proto.Message { return new(apihub.Lifecycle) },
	"google.cloud.apigeeregistry.v1.apihub.FieldSet":             func() proto.Message { return new(apihub.FieldSet) },
	"google.cloud.apigeeregistry.v1.apihub.FieldSetDefinition":   func() proto.Message { return new(apihub.FieldSetDefinition) },
	"google.cloud.apigeeregistry.v1.apihub.ReferenceList":        func() proto.Message { return new(apihub.ReferenceList) },
	"google.cloud.apigeeregistry.v1.apihub.TaxonomyList":         func() proto.Message { return new(apihub.TaxonomyList) },
	"google.cloud.apigeeregistry.v1.controller.Manifest":         func() proto.Message { return new(controller.Manifest) },
	"google.cloud.apigeeregistry.v1.controller.Receipt":          func() proto.Message { return new(controller.Receipt) },
	"google.cloud.apigeeregistry.v1.scoring.Score":               func() proto.Message { return new(scoring.Score) },
	"google.cloud.apigeeregistry.v1.scoring.ScoreDefinition":     func() proto.Message { return new(scoring.ScoreDefinition) },
	"google.cloud.apigeeregistry.v1.scoring.ScoreCard":           func() proto.Message { return new(scoring.ScoreCard) },
	"google.cloud.apigeeregistry.v1.scoring.ScoreCardDefinition": func() proto.Message { return new(scoring.ScoreCardDefinition) },
	"google.cloud.apigeeregistry.v1.style.StyleGuide":            func() proto.Message { return new(style.StyleGuide) },
	"google.cloud.apigeeregistry.v1.style.ConformanceReport":     func() proto.Message { return new(style.ConformanceReport) },
	"google.cloud.apigeeregistry.v1.style.Lint":                  func() proto.Message { return new(style.Lint) },
	"gnostic.metrics.Complexity":                                 func() proto.Message { return new(metrics.Complexity) },
	"gnostic.metrics.Vocabulary":                                 func() proto.Message { return new(metrics.Vocabulary) },
}
