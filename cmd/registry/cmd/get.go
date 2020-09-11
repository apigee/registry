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

package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	metrics "github.com/googleapis/gnostic/metrics"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var getContents bool

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVar(&getContents, "contents", false, "Get item contents (if applicable).")
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get entity values.",
	Long:  `Get entity values.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		var name string
		if len(args) > 0 {
			name = args[0]
		}

		if m := names.ProjectRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err = getProject(ctx, client, m[0], printProjectDetail)
		} else if m := names.ApiRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err = getAPI(ctx, client, m[0], printAPIDetail)
		} else if m := names.VersionRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err = getVersion(ctx, client, m[0], printVersionDetail)
		} else if m := names.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err = getSpec(ctx, client, m[0], getContents, printSpecDetail)
		} else if m := names.PropertyRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err = getProperty(ctx, client, m[0], printPropertyDetail)
		} else {
			log.Printf("Unsupported entity %+v", args)
		}
		if err != nil {
			log.Printf("%s", err.Error())
		}
	},
}

func getNamedProperty(ctx context.Context, client *gapic.RegistryClient, projectID string, subject string, relation string) error {
	request := &rpc.ListPropertiesRequest{
		Parent: subject,
		Filter: fmt.Sprintf("property_id = \"%s\"", relation),
	}
	it := client.ListProperties(ctx, request)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		printProperty(property)
	}
	return nil
}

func printPropertyDetail(property *rpc.Property) {
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
		default:
			fmt.Printf("%+v", v.MessageValue)
		}
	default:
		fmt.Printf("Unsupported property type: %s %s %+v\n", property.Subject, property.Relation, property.Value)
	}
	fmt.Printf("\n")
}

func printProjectDetail(message *rpc.Project) {
	printMessage(message)
}

func printAPIDetail(message *rpc.Api) {
	printMessage(message)
}

func printVersionDetail(message *rpc.Version) {
	printMessage(message)
}

func printSpecDetail(message *rpc.Spec) {
	if getContents {
		os.Stdout.Write(message.GetContents())
	} else {
		printMessage(message)
	}
}

func printMessage(message proto.Message) {
	fmt.Println(protojson.Format(message))
}

func unmarshalAndPrint(value []byte, message proto.Message) {
	err := proto.Unmarshal(value, message)
	if err != nil {
		fmt.Printf("%+v", err)
	} else {
		printMessage(message)
	}
}
