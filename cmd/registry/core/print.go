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

	"github.com/apigee/registry/cmd/registry/types"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

func IsPrintableType(mimeType string) bool {
	if mimeType == "text/plain" {
		return true
	} else if strings.HasPrefix(mimeType, "application/yaml") {
		return true
	} else if strings.HasPrefix(mimeType, "application/json") {
		return true
	} else {
		return false
	}
}

func PrintArtifactContents(artifact *rpc.Artifact) error {
	if IsPrintableType(artifact.GetMimeType()) {
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
	message, err := types.MessageForMimeType(artifact.GetMimeType())
	if err != nil {
		return nil, err
	}
	return unmarshal(artifact.GetContents(), message)
}

func unmarshal(value []byte, message proto.Message) (proto.Message, error) {
	if err := proto.Unmarshal(value, message); err != nil {
		return nil, err
	}
	return message, nil
}
