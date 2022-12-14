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

func PrintProject(project *rpc.Project) error {
	fmt.Println(project.Name)
	return nil
}

func PrintProjectDetail(message *rpc.Project) error {
	PrintMessage(message)
	return nil
}

func PrintAPI(api *rpc.Api) error {
	fmt.Println(api.Name)
	return nil
}

func PrintAPIDetail(message *rpc.Api) error {
	PrintMessage(message)
	return nil
}

func PrintDeployment(deployment *rpc.ApiDeployment) error {
	fmt.Println(deployment.Name)
	return nil
}

func PrintDeploymentDetail(message *rpc.ApiDeployment) error {
	PrintMessage(message)
	return nil
}

func PrintVersion(version *rpc.ApiVersion) error {
	fmt.Println(version.Name)
	return nil
}

func PrintVersionDetail(message *rpc.ApiVersion) error {
	PrintMessage(message)
	return nil
}

func PrintSpec(spec *rpc.ApiSpec) error {
	fmt.Println(spec.Name)
	return nil
}

func PrintSpecDetail(message *rpc.ApiSpec) error {
	PrintMessage(message)
	return nil
}

func WriteSpecContents(message *rpc.ApiSpec) error {
	contents := message.GetContents()
	if strings.Contains(message.GetMimeType(), "+gzip") {
		contents, _ = GUnzippedBytes(contents)
	}
	os.Stdout.Write(contents)
	return nil
}

func PrintArtifact(artifact *rpc.Artifact) error {
	fmt.Println(artifact.Name)
	return nil
}

func PrintArtifactDetail(artifact *rpc.Artifact) error {
	PrintMessage(artifact)
	return nil
}

func WriteArtifactContents(artifact *rpc.Artifact) error {
	os.Stdout.Write(artifact.GetContents())
	return nil
}

func PrintArtifactContents(artifact *rpc.Artifact) error {
	if artifact.GetMimeType() == "text/plain" {
		fmt.Printf("%s\n", string(artifact.GetContents()))
		return nil
	}

	message, err := GetArtifactMessageContents(artifact)
	if err != nil {
		return err
	}
	PrintMessage(message)
	return nil
}

func PrintMessage(message proto.Message) {
	fmt.Println(protojson.Format(message))
}

func GetArtifactMessageContents(artifact *rpc.Artifact) (proto.Message, error) {
	messageType, err := MessageTypeForMimeType(artifact.GetMimeType())
	if err != nil {
		return nil, err
	}
	switch messageType {
	case "gnostic.metrics.Complexity":
		return unmarshal(artifact.GetContents(), &metrics.Complexity{})
	case "gnostic.metrics.Vocabulary":
		return unmarshal(artifact.GetContents(), &metrics.Vocabulary{})
	case "gnostic.metrics.VersionHistory":
		return unmarshal(artifact.GetContents(), &metrics.VersionHistory{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.Index":
		return unmarshal(artifact.GetContents(), &rpc.Index{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.Lint":
		return unmarshal(artifact.GetContents(), &rpc.Lint{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.LintStats":
		return unmarshal(artifact.GetContents(), &rpc.LintStats{})
	case "google.cloud.apigeeregistry.v1.apihub.ApiSpecExtensionList":
		return unmarshal(artifact.GetContents(), &rpc.ApiSpecExtensionList{})
	case "google.cloud.apigeeregistry.v1.apihub.DisplaySettings":
		return unmarshal(artifact.GetContents(), &rpc.DisplaySettings{})
	case "google.cloud.apigeeregistry.v1.apihub.Lifecycle":
		return unmarshal(artifact.GetContents(), &rpc.Lifecycle{})
	case "google.cloud.apigeeregistry.v1.apihub.ReferenceList":
		return unmarshal(artifact.GetContents(), &rpc.ReferenceList{})
	case "google.cloud.apigeeregistry.v1.apihub.TaxonomyList":
		return unmarshal(artifact.GetContents(), &rpc.TaxonomyList{})
	case "google.cloud.apigeeregistry.v1.controller.Manifest":
		return unmarshal(artifact.GetContents(), &rpc.Manifest{})
	case "google.cloud.apigeeregistry.v1.controller.Receipt":
		return unmarshal(artifact.GetContents(), &rpc.Receipt{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.References":
		return unmarshal(artifact.GetContents(), &rpc.References{})
	case "google.cloud.apigeeregistry.v1.scoring.Score":
		return unmarshal(artifact.GetContents(), &rpc.Score{})
	case "google.cloud.apigeeregistry.v1.scoring.ScoreCard":
		return unmarshal(artifact.GetContents(), &rpc.ScoreCard{})
	case "google.cloud.apigeeregistry.v1.scoring.ScoreDefinition":
		return unmarshal(artifact.GetContents(), &rpc.ScoreDefinition{})
	case "google.cloud.apigeeregistry.v1.scoring.ScoreCardDefinition":
		return unmarshal(artifact.GetContents(), &rpc.ScoreCardDefinition{})
	case "google.cloud.apigeeregistry.v1.style.ConformanceReport":
		return unmarshal(artifact.GetContents(), &rpc.ConformanceReport{})
	case "google.cloud.apigeeregistry.v1.style.StyleGuide":
		return unmarshal(artifact.GetContents(), &rpc.StyleGuide{})
	case "gnostic.openapiv2.Document":
		return unmarshal(artifact.GetContents(), &openapiv2.Document{})
	case "gnostic.openapiv3.Document":
		return unmarshal(artifact.GetContents(), &openapiv3.Document{})
	default:
		return nil, fmt.Errorf("unsupported message type: %s", messageType)
	}
}

func unmarshal(value []byte, message proto.Message) (proto.Message, error) {
	if err := proto.Unmarshal(value, message); err != nil {
		return nil, err
	}
	return message, nil
}
