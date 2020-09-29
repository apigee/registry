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

func PrintVersion(version *rpc.Version) {
	fmt.Println(version.Name)
}

func PrintVersionDetail(message *rpc.Version) {
	PrintMessage(message)
}

func PrintSpec(spec *rpc.Spec) {
	fmt.Println(spec.Name)
}

func PrintSpecDetail(message *rpc.Spec) {
	PrintMessage(message)
}

func PrintSpecContents(message *rpc.Spec) {
	os.Stdout.Write(message.GetContents())
}

func PrintProperty(property *rpc.Property) {
	fmt.Println(property.Name)
}

func PrintPropertyDetail(property *rpc.Property) {
	switch v := property.Value.(type) {
	case *rpc.Property_StringValue:
		fmt.Printf("%s", v.StringValue)
	case *rpc.Property_Int64Value:
		fmt.Printf("%d", v.Int64Value)
	case *rpc.Property_DoubleValue:
		fmt.Printf("%f", v.DoubleValue)
	case *rpc.Property_BoolValue:
		fmt.Printf("%t", v.BoolValue)
	case *rpc.Property_BytesValue:
		fmt.Printf("%+v", v.BytesValue)
	case *rpc.Property_MessageValue:
		switch v.MessageValue.TypeUrl {
		case "gnostic.metrics.Complexity":
			unmarshalAndPrint(v.MessageValue.Value, &metrics.Complexity{})
		case "gnostic.metrics.Vocabulary":
			unmarshalAndPrint(v.MessageValue.Value, &metrics.Vocabulary{})
		case "google.cloud.apigee.registry.v1alpha1.Index":
			unmarshalAndPrint(v.MessageValue.Value, &rpc.Index{})
		case "gnostic.openapiv2.Document":
			unmarshalAndPrint(v.MessageValue.Value, &openapiv2.Document{})
		case "gnostic.openapiv3.Document":
			unmarshalAndPrint(v.MessageValue.Value, &openapiv3.Document{})
		default:
			fmt.Printf("%+v", v.MessageValue)
		}
	default:
		fmt.Printf("Unsupported property type: %s %s %+v\n", property.Subject, property.Relation, property.Value)
	}
	fmt.Printf("\n")
}

func PrintLabel(label *rpc.Label) {
	fmt.Println(label.Name)
}

func PrintLabelDetail(label *rpc.Label) {
	PrintMessage(label)
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
