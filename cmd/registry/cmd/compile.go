package cmd

import (
	"bytes"
	"compress/gzip"
	"context"
	"log"
	"strings"

	"apigov.dev/registry/cmd/registry/connection"
	"apigov.dev/registry/gapic"
	"apigov.dev/registry/models"
	rpc "apigov.dev/registry/rpc"
	rpcpb "apigov.dev/registry/rpc"
	"github.com/googleapis/gnostic/compiler"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Generate a compiled representation of an API spec",
	Long:  `Generate a compiled representation of an API spec.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			return
		}
		client, err := connection.NewClient()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ctx := context.TODO()
		name := args[0]
		if m := models.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err := compileSpec(ctx, client, m[0])
			if err != nil {
				log.Printf("%s", err.Error())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(compileCmd)
}

// ParentNameFromResourceName returns the name of a resource's parent.
func ParentNameFromResourceName(name string) string {
	parts := strings.Split(name, "/")
	return strings.Join(parts[0:len(parts)-2], "/")
}

func compileSpec(ctx context.Context,
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
		err = uploadBytesForSpec(ctx, client, ParentNameFromResourceName(spec.GetName()), "swagger.pb", spec.GetStyle(), document)
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
		err = uploadBytesForSpec(ctx, client, ParentNameFromResourceName(spec.GetName()), "openapi.pb", spec.GetStyle(), document)
		if err != nil {
			return err
		}
	}
	return nil
}

func uploadBytesForSpec(ctx context.Context, client connection.Client, parent string, specID string, style string, document proto.Message) error {
	// gzip the spec before uploading it
	messageData, err := proto.Marshal(document)
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, err = zw.Write(messageData)
	if err != nil {
		log.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
	request := &rpc.CreateSpecRequest{}
	request.Parent = parent
	request.SpecId = specID
	request.Spec = &rpcpb.Spec{}
	request.Spec.Style = style
	request.Spec.Contents = buf.Bytes()
	_, err = client.CreateSpec(ctx, request)
	if err != nil {
		// if this fails, we should try calling UpdateSpec
		return err
	}
	return nil
}
