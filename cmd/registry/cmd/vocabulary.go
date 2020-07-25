package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"apigov.dev/registry/connection"
	"apigov.dev/registry/gapic"
	"apigov.dev/registry/server/models"
	rpc "apigov.dev/registry/rpc"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/googleapis/gnostic/compiler"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

// vocabularyCmd represents the vocabulary command
var vocabularyCmd = &cobra.Command{
	Use:   "vocabulary",
	Short: "Generate a summary of an API spec",
	Long:  `Generate a summary of an API spec.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		log.Printf("vocabulary called %+v", args)
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		name := args[0]
		if m := models.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				// iterate through a collection of specs and vocabulary each
				completions := make(chan int)
				processes := 0
				err = listSpecs(ctx, client, segments, func(spec *rpc.Spec) {
					fmt.Println(spec.Name)
					m := models.SpecRegexp().FindAllStringSubmatch(spec.Name, -1)
					if m != nil {
						processes++
						go func() {
							vocabularySpec(ctx, client, m[0])
							completions <- 1
						}()
					}
				})
				for i := 0; i < processes; i++ {
					<-completions
				}

			} else {
				err := vocabularySpec(ctx, client, segments)
				if err != nil {
					log.Printf("%s", err.Error())
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(vocabularyCmd)
}

func vocabularySpec(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {

	name := resourceNameOfSpec(segments[1:])
	request := &rpc.GetSpecRequest{
		Name: name,
		View: rpc.SpecView_FULL,
	}
	spec, err := client.GetSpec(ctx, request)
	if err != nil {
		return err
	}

	log.Printf("computing vocabulary of %s", spec.Name)
	if strings.HasPrefix(spec.GetStyle(), "openapi/v2") {
		data, err := getBytesForSpec(spec)
		if err != nil {
			return nil
		}
		info, err := compiler.ReadInfoFromBytes(spec.GetName(), data)
		if err != nil {
			return err
		}
		document, err := openapi_v2.NewDocument(info, compiler.NewContextWithExtensions("$root", nil, nil))
		if err != nil {
			return err
		}
		log.Printf("%+v", document)

		vocabulary := processDocumentV2(document)
		log.Printf("%+v", vocabulary)

		projectID := segments[1]
		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "vocabulary"
		messageData, err := proto.Marshal(vocabulary)
		anyValue := &any.Any{
			TypeUrl: "Vocabulary",
			Value:   messageData,
		}
		property.Value = &rpc.Property_MessageValue{MessageValue: anyValue}
		err = setProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}

	}
	if strings.HasPrefix(spec.GetStyle(), "openapi/v3") {
		data, err := getBytesForSpec(spec)
		if err != nil {
			return nil
		}
		info, err := compiler.ReadInfoFromBytes(spec.GetName(), data)
		if err != nil {
			return err
		}
		document, err := openapi_v3.NewDocument(info, compiler.NewContextWithExtensions("$root", nil, nil))
		if err != nil {
			return err
		}
		vocabulary := processDocumentV3(document)

		projectID := segments[1]

		log.Printf("%s", projectID)
		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "vocabulary"
		messageData, err := proto.Marshal(vocabulary)
		anyValue := &any.Any{
			TypeUrl: "Vocabulary",
			Value:   messageData,
		}
		property.Value = &rpc.Property_MessageValue{MessageValue: anyValue}
		err = setProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}

	}
	return nil
}
