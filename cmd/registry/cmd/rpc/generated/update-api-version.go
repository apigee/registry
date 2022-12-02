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

var UpdateApiVersionInput rpcpb.UpdateApiVersionRequest

var UpdateApiVersionFromFile string

var UpdateApiVersionInputApiVersionLabels []string

var UpdateApiVersionInputApiVersionAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(UpdateApiVersionCmd)

	UpdateApiVersionInput.ApiVersion = new(rpcpb.ApiVersion)

	UpdateApiVersionInput.UpdateMask = new(fieldmaskpb.FieldMask)

	UpdateApiVersionCmd.Flags().StringVar(&UpdateApiVersionInput.ApiVersion.Name, "api_version.name", "", "Resource name.")

	UpdateApiVersionCmd.Flags().StringVar(&UpdateApiVersionInput.ApiVersion.DisplayName, "api_version.display_name", "", "Human-meaningful name.")

	UpdateApiVersionCmd.Flags().StringVar(&UpdateApiVersionInput.ApiVersion.Description, "api_version.description", "", "A detailed description.")

	UpdateApiVersionCmd.Flags().StringVar(&UpdateApiVersionInput.ApiVersion.State, "api_version.state", "", "A user-definable description of the lifecycle...")

	UpdateApiVersionCmd.Flags().StringArrayVar(&UpdateApiVersionInputApiVersionLabels, "api_version.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	UpdateApiVersionCmd.Flags().StringArrayVar(&UpdateApiVersionInputApiVersionAnnotations, "api_version.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	UpdateApiVersionCmd.Flags().StringVar(&UpdateApiVersionInput.ApiVersion.PrimarySpec, "api_version.primary_spec", "", "The primary spec for this version.  Format:...")

	UpdateApiVersionCmd.Flags().StringSliceVar(&UpdateApiVersionInput.UpdateMask.Paths, "update_mask.paths", []string{}, "The set of field mask paths.")

	UpdateApiVersionCmd.Flags().BoolVar(&UpdateApiVersionInput.AllowMissing, "allow_missing", false, "If set to true, and the version is not found, a...")

	UpdateApiVersionCmd.Flags().StringVar(&UpdateApiVersionFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var UpdateApiVersionCmd = &cobra.Command{
	Use:   "update-api-version",
	Short: "UpdateApiVersion can be used to modify a...",
	Long:  "UpdateApiVersion can be used to modify a specified version.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if UpdateApiVersionFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if UpdateApiVersionFromFile != "" {
			in, err = os.Open(UpdateApiVersionFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &UpdateApiVersionInput)
			if err != nil {
				return err
			}

		}

		if len(UpdateApiVersionInputApiVersionLabels) > 0 {
			UpdateApiVersionInput.ApiVersion.Labels = make(map[string]string)
		}
		for _, item := range UpdateApiVersionInputApiVersionLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			UpdateApiVersionInput.ApiVersion.Labels[split[0]] = split[1]
		}

		if len(UpdateApiVersionInputApiVersionAnnotations) > 0 {
			UpdateApiVersionInput.ApiVersion.Annotations = make(map[string]string)
		}
		for _, item := range UpdateApiVersionInputApiVersionAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			UpdateApiVersionInput.ApiVersion.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "UpdateApiVersion", &UpdateApiVersionInput)
		}
		resp, err := RegistryClient.UpdateApiVersion(ctx, &UpdateApiVersionInput)
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
