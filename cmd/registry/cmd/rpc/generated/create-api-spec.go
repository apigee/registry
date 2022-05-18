// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"

	"strings"
)

var CreateApiSpecInput rpcpb.CreateApiSpecRequest

var CreateApiSpecFromFile string

var CreateApiSpecInputApiSpecLabels []string

var CreateApiSpecInputApiSpecAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(CreateApiSpecCmd)

	CreateApiSpecInput.ApiSpec = new(rpcpb.ApiSpec)

	CreateApiSpecCmd.Flags().StringVar(&CreateApiSpecInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	CreateApiSpecCmd.Flags().StringVar(&CreateApiSpecInput.ApiSpec.Name, "api_spec.name", "", "Resource name.")

	CreateApiSpecCmd.Flags().StringVar(&CreateApiSpecInput.ApiSpec.Filename, "api_spec.filename", "", "A possibly-hierarchical name used to refer to the...")

	CreateApiSpecCmd.Flags().StringVar(&CreateApiSpecInput.ApiSpec.Description, "api_spec.description", "", "A detailed description.")

	CreateApiSpecCmd.Flags().StringVar(&CreateApiSpecInput.ApiSpec.MimeType, "api_spec.mime_type", "", "A style (format) descriptor for this spec that is...")

	CreateApiSpecCmd.Flags().StringVar(&CreateApiSpecInput.ApiSpec.SourceUri, "api_spec.source_uri", "", "The original source URI of the spec (if one...")

	CreateApiSpecCmd.Flags().BytesHexVar(&CreateApiSpecInput.ApiSpec.Contents, "api_spec.contents", []byte{}, "Input only. The contents of the spec.  Provided...")

	CreateApiSpecCmd.Flags().StringArrayVar(&CreateApiSpecInputApiSpecLabels, "api_spec.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	CreateApiSpecCmd.Flags().StringArrayVar(&CreateApiSpecInputApiSpecAnnotations, "api_spec.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	CreateApiSpecCmd.Flags().StringVar(&CreateApiSpecInput.ApiSpecId, "api_spec_id", "", "Required. The ID to use for the spec, which will...")

	CreateApiSpecCmd.Flags().StringVar(&CreateApiSpecFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var CreateApiSpecCmd = &cobra.Command{
	Use:   "create-api-spec",
	Short: "CreateApiSpec creates a specified spec.",
	Long:  "CreateApiSpec creates a specified spec.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if CreateApiSpecFromFile == "" {

			cmd.MarkFlagRequired("parent")

			cmd.MarkFlagRequired("api_spec_id")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if CreateApiSpecFromFile != "" {
			in, err = os.Open(CreateApiSpecFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &CreateApiSpecInput)
			if err != nil {
				return err
			}

		}

		if len(CreateApiSpecInputApiSpecLabels) > 0 {
			CreateApiSpecInput.ApiSpec.Labels = make(map[string]string)
		}
		for _, item := range CreateApiSpecInputApiSpecLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateApiSpecInput.ApiSpec.Labels[split[0]] = split[1]
		}

		if len(CreateApiSpecInputApiSpecAnnotations) > 0 {
			CreateApiSpecInput.ApiSpec.Annotations = make(map[string]string)
		}
		for _, item := range CreateApiSpecInputApiSpecAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateApiSpecInput.ApiSpec.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "CreateApiSpec", &CreateApiSpecInput)
		}
		resp, err := RegistryClient.CreateApiSpec(ctx, &CreateApiSpecInput)
		if err != nil {
			return err
		}

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(resp)

		return err
	},
}
