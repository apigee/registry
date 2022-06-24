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

var UpdateApiInput rpcpb.UpdateApiRequest

var UpdateApiFromFile string

var UpdateApiInputApiLabels []string

var UpdateApiInputApiAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(UpdateApiCmd)

	UpdateApiInput.Api = new(rpcpb.Api)

	UpdateApiInput.UpdateMask = new(fieldmaskpb.FieldMask)

	UpdateApiCmd.Flags().StringVar(&UpdateApiInput.Api.Name, "api.name", "", "Resource name.")

	UpdateApiCmd.Flags().StringVar(&UpdateApiInput.Api.DisplayName, "api.display_name", "", "Human-meaningful name.")

	UpdateApiCmd.Flags().StringVar(&UpdateApiInput.Api.Description, "api.description", "", "A detailed description.")

	UpdateApiCmd.Flags().StringVar(&UpdateApiInput.Api.Availability, "api.availability", "", "A user-definable description of the availability...")

	UpdateApiCmd.Flags().StringVar(&UpdateApiInput.Api.RecommendedVersion, "api.recommended_version", "", "The recommended version of the API.  Format:...")

	UpdateApiCmd.Flags().StringVar(&UpdateApiInput.Api.RecommendedDeployment, "api.recommended_deployment", "", "The recommended deployment of the API.  Format:...")

	UpdateApiCmd.Flags().StringArrayVar(&UpdateApiInputApiLabels, "api.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	UpdateApiCmd.Flags().StringArrayVar(&UpdateApiInputApiAnnotations, "api.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	UpdateApiCmd.Flags().StringSliceVar(&UpdateApiInput.UpdateMask.Paths, "update_mask.paths", []string{}, "The set of field mask paths.")

	UpdateApiCmd.Flags().BoolVar(&UpdateApiInput.AllowMissing, "allow_missing", false, "If set to true, and the api is not found, a new...")

	UpdateApiCmd.Flags().StringVar(&UpdateApiFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var UpdateApiCmd = &cobra.Command{
	Use:   "update-api",
	Short: "UpdateApi can be used to modify a specified API.",
	Long:  "UpdateApi can be used to modify a specified API.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if UpdateApiFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if UpdateApiFromFile != "" {
			in, err = os.Open(UpdateApiFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &UpdateApiInput)
			if err != nil {
				return err
			}

		}

		if len(UpdateApiInputApiLabels) > 0 {
			UpdateApiInput.Api.Labels = make(map[string]string)
		}
		for _, item := range UpdateApiInputApiLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			UpdateApiInput.Api.Labels[split[0]] = split[1]
		}

		if len(UpdateApiInputApiAnnotations) > 0 {
			UpdateApiInput.Api.Annotations = make(map[string]string)
		}
		for _, item := range UpdateApiInputApiAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			UpdateApiInput.Api.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "UpdateApi", &UpdateApiInput)
		}
		resp, err := RegistryClient.UpdateApi(ctx, &UpdateApiInput)
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
