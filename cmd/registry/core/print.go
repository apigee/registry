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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	metrics "github.com/google/gnostic/metrics"
	openapiv2 "github.com/google/gnostic/openapiv2"
	openapiv3 "github.com/google/gnostic/openapiv3"
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

func PrintDeployment(deployment *rpc.ApiDeployment) {
	fmt.Println(deployment.Name)
}

func PrintDeploymentDetail(message *rpc.ApiDeployment) {
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
	if strings.Contains(message.GetMimeType(), "+gzip") {
		contents, _ = GUnzippedBytes(contents)
	}
	os.Stdout.Write(contents)
}

func PrintArtifact(artifact *rpc.Artifact) {
	fmt.Println(artifact.Name)
}

func PrintArtifactDetail(artifact *rpc.Artifact) {
	PrintMessage(artifact)
}

func PrintArtifactContents(artifact *rpc.Artifact) {
	if artifact.GetMimeType() == "text/plain" {
		fmt.Printf("%s\n", string(artifact.GetContents()))
		return
	}
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
	case "google.cloud.apigeeregistry.applications.v1alpha1.ConformanceReport":
		unmarshalAndPrint(artifact.GetContents(), &rpc.ConformanceReport{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.Index":
		unmarshalAndPrint(artifact.GetContents(), &rpc.Index{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.Lint":
		unmarshalAndPrint(artifact.GetContents(), &rpc.Lint{})
	case "google.cloud.apigeeregistry.v1.apihub.DisplaySettings":
		unmarshalAndPrint(artifact.GetContents(), &rpc.DisplaySettings{})
	case "google.cloud.apigeeregistry.v1.apihub.Lifecycle":
		unmarshalAndPrint(artifact.GetContents(), &rpc.Lifecycle{})
	case "google.cloud.apigeeregistry.v1.apihub.ReferenceList":
		unmarshalAndPrint(artifact.GetContents(), &rpc.ReferenceList{})
	case "google.cloud.apigeeregistry.v1.apihub.TaxonomyList":
		unmarshalAndPrint(artifact.GetContents(), &rpc.TaxonomyList{})
	case "google.cloud.apigeeregistry.v1.controller.Manifest":
		unmarshalAndPrint(artifact.GetContents(), &rpc.Manifest{})
	case "google.cloud.apigeeregistry.v1.controller.Receipt":
		unmarshalAndPrint(artifact.GetContents(), &rpc.Receipt{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.References":
		unmarshalAndPrint(artifact.GetContents(), &rpc.References{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.StyleGuide":
		unmarshalAndPrint(artifact.GetContents(), &rpc.StyleGuide{})
	case "gnostic.openapiv2.Document":
		unmarshalAndPrint(artifact.GetContents(), &openapiv2.Document{})
	case "gnostic.openapiv3.Document":
		unmarshalAndPrint(artifact.GetContents(), &openapiv3.Document{})
	default:
		fmt.Printf("%+v", artifact.GetContents())
	}
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
