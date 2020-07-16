package cmd

import (
	"context"
	"fmt"
	"log"

	"apigov.dev/registry/cmd/registry/connection"
	"apigov.dev/registry/gapic"
	"apigov.dev/registry/models"
	rpc "apigov.dev/registry/rpc"
	metrics "github.com/googleapis/gnostic/metrics"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get property values.",
	Long:  `Get property values.`,
	Run: func(cmd *cobra.Command, args []string) {

		client, err := connection.NewClient()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ctx := context.TODO()

		var name, property string
		if len(args) > 0 {
			name = args[0]
		}
		if len(args) > 1 {
			property = args[1]
		}

		// first look for the main resource types

		if m := models.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			log.Printf(" get a spec")

			_, err = getSpec(ctx, client, m[0], printSpecDetail)

		} else if m := models.ProjectRegexp().FindAllStringSubmatch(name, -1); m != nil {
			// find all matching properties for a project
			segments := m[0]
			err = getNamedProperty(ctx, client, segments[1], "", property)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		} else if m := models.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			// find all matching properties for matching specs
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listSpecs(ctx, client, segments, func(spec *rpc.Spec) {
					getNamedProperty(ctx, client, segments[1], spec.GetName(), property)
				})
				if err != nil {
					log.Fatalf("%s", err.Error())
				}
			} else {
				spec, err := getSpec(ctx, client, segments, func(s *rpc.Spec) {})
				if err != nil {
					log.Fatalf("%s", err.Error())
				}
				err = getNamedProperty(ctx, client, segments[1], spec.GetName(), property)
				if err != nil {
					log.Fatalf("%s", err.Error())
				}
			}
		} else {
			log.Printf("Unable to process %+v", args)
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

func init() {
	rootCmd.AddCommand(getCmd)
}

func printPropertyDetail(property *rpc.Property) {
	fmt.Printf("%s %s %+v\n", property.Subject, property.Relation, property.Value)
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
		messageType := v.MessageValue.TypeUrl
		if messageType == "Complexity" {
			var msg metrics.Complexity
			err := proto.Unmarshal(v.MessageValue.Value, &msg)
			if err != nil {
				fmt.Printf("%+v", err)
			} else {
				fmt.Printf("%+v", &msg)
			}
		} else {
			fmt.Printf("%+v", v.MessageValue)
		}
	default:

	}
	fmt.Printf("\n")
}

func printSpecDetail(spec *rpc.Spec) {
	fmt.Println(spec.Name)
	fmt.Printf("%+v\n", spec)
}
