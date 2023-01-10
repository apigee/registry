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
	"strings"

	"github.com/apigee/registry/cmd/registry/types"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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
