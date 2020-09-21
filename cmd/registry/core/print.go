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
	printMessage(message)
}

func PrintAPI(api *rpc.Api) {
	fmt.Println(api.Name)
}

func PrintAPIDetail(message *rpc.Api) {
	printMessage(message)
}

func PrintVersion(version *rpc.Version) {
	fmt.Println(version.Name)
}

func PrintVersionDetail(message *rpc.Version) {
	printMessage(message)
}

func PrintSpec(spec *rpc.Spec) {
	fmt.Println(spec.Name)
}

func PrintSpecDetail(message *rpc.Spec) {
	printMessage(message)
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
	printMessage(label)
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
