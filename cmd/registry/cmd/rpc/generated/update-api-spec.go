// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"

	"strings"
)

var UpdateApiSpecInput rpcpb.UpdateApiSpecRequest

var UpdateApiSpecFromFile string

var UpdateApiSpecInputApiSpecLabels []string

var UpdateApiSpecInputApiSpecAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(UpdateApiSpecCmd)

	UpdateApiSpecInput.ApiSpec = new(rpcpb.ApiSpec)

	UpdateApiSpecInput.UpdateMask = new(fieldmaskpb.FieldMask)

	UpdateApiSpecCmd.Flags().StringVar(&UpdateApiSpecInput.ApiSpec.Name, "api_spec.name", "", "Resource name.")

	UpdateApiSpecCmd.Flags().StringVar(&UpdateApiSpecInput.ApiSpec.Filename, "api_spec.filename", "", "A possibly-hierarchical name used to refer to the...")

	UpdateApiSpecCmd.Flags().StringVar(&UpdateApiSpecInput.ApiSpec.Description, "api_spec.description", "", "A detailed description.")

	UpdateApiSpecCmd.Flags().StringVar(&UpdateApiSpecInput.ApiSpec.MimeType, "api_spec.mime_type", "", "A style (format) descriptor for this spec that is...")

	UpdateApiSpecCmd.Flags().StringVar(&UpdateApiSpecInput.ApiSpec.SourceUri, "api_spec.source_uri", "", "The original source URI of the spec (if one...")

	UpdateApiSpecCmd.Flags().BytesHexVar(&UpdateApiSpecInput.ApiSpec.Contents, "api_spec.contents", []byte{}, "Input only. The contents of the spec.  Provided...")

	UpdateApiSpecCmd.Flags().StringArrayVar(&UpdateApiSpecInputApiSpecLabels, "api_spec.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	UpdateApiSpecCmd.Flags().StringArrayVar(&UpdateApiSpecInputApiSpecAnnotations, "api_spec.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	UpdateApiSpecCmd.Flags().StringSliceVar(&UpdateApiSpecInput.UpdateMask.Paths, "update_mask.paths", []string{}, "The set of field mask paths.")

	UpdateApiSpecCmd.Flags().BoolVar(&UpdateApiSpecInput.AllowMissing, "allow_missing", false, "If set to true, and the spec is not found, a new...")

	UpdateApiSpecCmd.Flags().StringVar(&UpdateApiSpecFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var UpdateApiSpecCmd = &cobra.Command{
	Use:   "update-api-spec",
	Short: "UpdateApiSpec can be used to modify a specified...",
	Long:  "UpdateApiSpec can be used to modify a specified spec.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if UpdateApiSpecFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if UpdateApiSpecFromFile != "" {
			in, err = os.Open(UpdateApiSpecFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &UpdateApiSpecInput)
			if err != nil {
				return err
			}

		}

		if len(UpdateApiSpecInputApiSpecLabels) > 0 {
			UpdateApiSpecInput.ApiSpec.Labels = make(map[string]string)
		}
		for _, item := range UpdateApiSpecInputApiSpecLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			UpdateApiSpecInput.ApiSpec.Labels[split[0]] = split[1]
		}

		if len(UpdateApiSpecInputApiSpecAnnotations) > 0 {
			UpdateApiSpecInput.ApiSpec.Annotations = make(map[string]string)
		}
		for _, item := range UpdateApiSpecInputApiSpecAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			UpdateApiSpecInput.ApiSpec.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "UpdateApiSpec", &UpdateApiSpecInput)
		}
		resp, err := RegistryClient.UpdateApiSpec(ctx, &UpdateApiSpecInput)
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
