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

var UpdateApiDeploymentInput rpcpb.UpdateApiDeploymentRequest

var UpdateApiDeploymentFromFile string

var UpdateApiDeploymentInputApiDeploymentLabels []string

var UpdateApiDeploymentInputApiDeploymentAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(UpdateApiDeploymentCmd)

	UpdateApiDeploymentInput.ApiDeployment = new(rpcpb.ApiDeployment)

	UpdateApiDeploymentInput.UpdateMask = new(fieldmaskpb.FieldMask)

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentInput.ApiDeployment.Name, "api_deployment.name", "", "Resource name.")

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentInput.ApiDeployment.DisplayName, "api_deployment.display_name", "", "Human-meaningful name.")

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentInput.ApiDeployment.Description, "api_deployment.description", "", "A detailed description.")

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentInput.ApiDeployment.ApiSpecRevision, "api_deployment.api_spec_revision", "", "The full resource name (including revision id) of...")

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentInput.ApiDeployment.EndpointUri, "api_deployment.endpoint_uri", "", "The address where the deployment is serving....")

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentInput.ApiDeployment.ExternalChannelUri, "api_deployment.external_channel_uri", "", "The address of the external channel of the API...")

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentInput.ApiDeployment.IntendedAudience, "api_deployment.intended_audience", "", "Text briefly identifying the intended audience of...")

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentInput.ApiDeployment.AccessGuidance, "api_deployment.access_guidance", "", "Text briefly describing how to access the...")

	UpdateApiDeploymentCmd.Flags().StringArrayVar(&UpdateApiDeploymentInputApiDeploymentLabels, "api_deployment.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	UpdateApiDeploymentCmd.Flags().StringArrayVar(&UpdateApiDeploymentInputApiDeploymentAnnotations, "api_deployment.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	UpdateApiDeploymentCmd.Flags().StringSliceVar(&UpdateApiDeploymentInput.UpdateMask.Paths, "update_mask.paths", []string{}, "The set of field mask paths.")

	UpdateApiDeploymentCmd.Flags().BoolVar(&UpdateApiDeploymentInput.AllowMissing, "allow_missing", false, "If set to true, and the deployment is not found,...")

	UpdateApiDeploymentCmd.Flags().StringVar(&UpdateApiDeploymentFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var UpdateApiDeploymentCmd = &cobra.Command{
	Use:   "update-api-deployment",
	Short: "UpdateApiDeployment can be used to modify a...",
	Long:  "UpdateApiDeployment can be used to modify a specified deployment.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if UpdateApiDeploymentFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if UpdateApiDeploymentFromFile != "" {
			in, err = os.Open(UpdateApiDeploymentFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &UpdateApiDeploymentInput)
			if err != nil {
				return err
			}

		}

		if len(UpdateApiDeploymentInputApiDeploymentLabels) > 0 {
			UpdateApiDeploymentInput.ApiDeployment.Labels = make(map[string]string)
		}
		for _, item := range UpdateApiDeploymentInputApiDeploymentLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			UpdateApiDeploymentInput.ApiDeployment.Labels[split[0]] = split[1]
		}

		if len(UpdateApiDeploymentInputApiDeploymentAnnotations) > 0 {
			UpdateApiDeploymentInput.ApiDeployment.Annotations = make(map[string]string)
		}
		for _, item := range UpdateApiDeploymentInputApiDeploymentAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			UpdateApiDeploymentInput.ApiDeployment.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "UpdateApiDeployment", &UpdateApiDeploymentInput)
		}
		resp, err := RegistryClient.UpdateApiDeployment(ctx, &UpdateApiDeploymentInput)
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
