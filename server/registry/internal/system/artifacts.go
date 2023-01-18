// Copyright 2023 Google LLC. All Rights Reserved.
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

package system

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/buildinfo"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

func Reserved(name names.Artifact) bool {
	return name.ApiID() == "" && strings.HasPrefix(name.ArtifactID(), "system-")
}

func Artifact(name names.Artifact) (*rpc.Artifact, error) {
	contents, mimeType, err := contents(name)
	if err != nil {
		return nil, err
	}
	message := &rpc.Artifact{
		Name:      name.String(),
		MimeType:  mimeType,
		SizeBytes: int32(len(contents)),
		Hash:      hashForBytes(contents),
	}
	return message, nil
}

func ArtifactContents(name names.Artifact) (*httpbody.HttpBody, error) {
	contents, mimeType, err := contents(name)
	if err != nil {
		return nil, err
	}
	return &httpbody.HttpBody{
		ContentType: mimeType,
		Data:        contents,
	}, nil
}

func contents(name names.Artifact) ([]byte, string, error) {
	switch name.ArtifactID() {
	case "system-buildinfo":
		return systemBuildinfo()
	default:
		return nil, "", status.Errorf(codes.NotFound, "%q not found", name)
	}
}

func systemBuildinfo() ([]byte, string, error) {
	status := buildinfo.BuildInfo()
	// Marshal the artifact content as JSON using the protobuf marshaller.
	var s []byte
	s, err := protojson.MarshalOptions{
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
		Indent:          "  ",
		UseProtoNames:   false,
	}.Marshal(status)
	if err != nil {
		return nil, "", err
	}
	// Unmarshal the JSON with yaml.v3 so that we can re-marshal it as YAML.
	var doc yaml.Node
	err = yaml.Unmarshal([]byte(s), &doc)
	if err != nil {
		return nil, "", err
	}
	// The top-level node is a "document" node. We need to marshal the node below it.
	node := doc.Content[0]
	// Restyle the YAML representation so that it will be serialized with YAML defaults.
	styleForYAML(node)
	contents, err := encode(node)
	if err != nil {
		return nil, "", err
	}
	return contents, "application/yaml;type=BuildInfo", nil
}

// Prefer this encoder because it uses tighter 2-space indentation.
func yamlEncoder(dst io.Writer) *yaml.Encoder {
	enc := yaml.NewEncoder(dst)
	enc.SetIndent(2)
	return enc
}

// Encode a patch model.
func encode(model interface{}) ([]byte, error) {
	var b bytes.Buffer
	err := yamlEncoder(&b).Encode(model)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// styleForYAML sets the style field on a tree of yaml.Nodes for YAML export.
func styleForYAML(node *yaml.Node) {
	node.Style = 0
	for _, n := range node.Content {
		styleForYAML(n)
	}
}

func hashForBytes(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	h := sha256.New()
	_, _ = h.Write(b)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
