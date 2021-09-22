// Copyright 2021 Google LLC. All Rights Reserved.
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

package core

import (
	"fmt"
	"regexp"
	"strings"
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

// TODO: tighten these up, possibly using regular expressions.

// IsAsyncAPIv2 returns true if a MIME type represents an AsyncAPI v2 spec.
func IsAsyncAPIv2(mimeType string) bool {
	return strings.Contains(mimeType, "asyncapi") &&
		strings.Contains(mimeType, "version=2")
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

// IsZipArchive returns true if a MIME type represents a type stored as a multifile Zip archive.
func IsZipArchive(mimeType string) bool {
	return strings.Contains(mimeType, "+zip")
}

// MimeTypeForMessageType returns a MIME type that represents a Protocol Buffer message type.
func MimeTypeForMessageType(protoType string) string {
	return fmt.Sprintf("application/octet-stream;type=%s", protoType)
}

// MessageTypeForMimeType returns the Protocol Buffer message type represented by a MIME type.
func MessageTypeForMimeType(protoType string) (string, error) {
	re := regexp.MustCompile("^application/octet-stream;type=(.*)$")
	m := re.FindStringSubmatch(protoType)
	if m == nil || len(m) < 2 || len(m[1]) == 0 {
		return "", fmt.Errorf("invalid Protocol Buffer type: %s", protoType)
	}
	return m[1], nil
}
