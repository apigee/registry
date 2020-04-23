package cmd

import (
	"context"
	"fmt"
	"log"

	"apigov.dev/flame/cmd/flame/connection"
	"apigov.dev/flame/gapic"
	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
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

		if m := models.ProjectRegexp().FindAllStringSubmatch(name, -1); m != nil {
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
				err = listSpecs(ctx, client, segments, func(spec *rpc.Spec) error {
					err = getNamedProperty(ctx, client, segments[1], spec.GetName(), property)
					return err
				})
				if err != nil {
					log.Fatalf("%s", err.Error())
				}
			} else {
				spec, err := getSpec(ctx, client, segments)
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

func getNamedProperty(ctx context.Context, client *gapic.FlameClient, projectID string, subject string, relation string) error {
	request := &rpc.ListPropertiesRequest{
		Parent:   "projects/" + projectID,
		Subject:  subject,
		Relation: relation,
	}
	it := client.ListProperties(ctx, request)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		fmt.Printf("%s %s %+v\n", property.Subject, property.Relation, property.Value)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(getCmd)
}
