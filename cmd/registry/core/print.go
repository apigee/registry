// Copyright 2020 Google LLC. All Rights Reserved.
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
	"os"
	"strings"

	"github.com/apigee/registry/rpc"
	metrics "github.com/googleapis/gnostic/metrics"
	openapiv2 "github.com/googleapis/gnostic/openapiv2"
	openapiv3 "github.com/googleapis/gnostic/openapiv3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func PrintProject(project *rpc.Project) {
	fmt.Println(project.Name)
}

func PrintProjectDetail(message *rpc.Project) {
	PrintMessage(message)
}

func PrintAPI(api *rpc.Api) {
	fmt.Println(api.Name)
}

func PrintAPIDetail(message *rpc.Api) {
	PrintMessage(message)
}

func PrintVersion(version *rpc.ApiVersion) {
	fmt.Println(version.Name)
}

func PrintVersionDetail(message *rpc.ApiVersion) {
	PrintMessage(message)
}

func PrintSpec(spec *rpc.ApiSpec) {
	fmt.Println(spec.Name)
}

func PrintSpecDetail(message *rpc.ApiSpec) {
	PrintMessage(message)
}

func PrintSpecContents(message *rpc.ApiSpec) {
	contents := message.GetContents()
	if strings.HasSuffix(message.GetMimeType(), "+gzip") {
		contents, _ = GUnzippedBytes(contents)
	}
	os.Stdout.Write(contents)
}

func PrintArtifact(artifact *rpc.Artifact) {
	fmt.Println(artifact.Name)
}

func PrintArtifactDetail(artifact *rpc.Artifact) {
	messageType, err := MessageTypeForMimeType(artifact.GetMimeType())
	if err != nil {
		fmt.Println(artifact.Name)
	}
	switch messageType {
	case "gnostic.metrics.Complexity":
		unmarshalAndPrint(artifact.GetContents(), &metrics.Complexity{})
	case "gnostic.metrics.Vocabulary":
		unmarshalAndPrint(artifact.GetContents(), &metrics.Vocabulary{})
	case "gnostic.metrics.VersionHistory":
		unmarshalAndPrint(artifact.GetContents(), &metrics.VersionHistory{})
	case "google.cloud.apigee.registry.applications.v1alpha1.Index":
		unmarshalAndPrint(artifact.GetContents(), &rpc.Index{})
	case "gnostic.openapiv2.Document":
		unmarshalAndPrint(artifact.GetContents(), &openapiv2.Document{})
	case "gnostic.openapiv3.Document":
		unmarshalAndPrint(artifact.GetContents(), &openapiv3.Document{})
	default:
		fmt.Printf("%+v", artifact.GetContents())
	}
}

func PrintArtifactContents(message *rpc.Artifact) {
	PrintArtifactDetail(message)
}

func PrintMessage(message proto.Message) {
	fmt.Println(protojson.Format(message))
}

func unmarshalAndPrint(value []byte, message proto.Message) {
	err := proto.Unmarshal(value, message)
	if err != nil {
		fmt.Printf("%+v", err)
	} else {
		PrintMessage(message)
	}
}
