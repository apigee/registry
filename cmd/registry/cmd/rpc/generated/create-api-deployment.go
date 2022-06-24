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

var CreateApiDeploymentInput rpcpb.CreateApiDeploymentRequest

var CreateApiDeploymentFromFile string

var CreateApiDeploymentInputApiDeploymentLabels []string

var CreateApiDeploymentInputApiDeploymentAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(CreateApiDeploymentCmd)

	CreateApiDeploymentInput.ApiDeployment = new(rpcpb.ApiDeployment)

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeployment.Name, "api_deployment.name", "", "Resource name.")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeployment.DisplayName, "api_deployment.display_name", "", "Human-meaningful name.")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeployment.Description, "api_deployment.description", "", "A detailed description.")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeployment.ApiSpecRevision, "api_deployment.api_spec_revision", "", "The full resource name (including revision id) of...")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeployment.EndpointUri, "api_deployment.endpoint_uri", "", "The address where the deployment is serving....")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeployment.ExternalChannelUri, "api_deployment.external_channel_uri", "", "The address of the external channel of the API...")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeployment.IntendedAudience, "api_deployment.intended_audience", "", "Text briefly identifying the intended audience of...")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeployment.AccessGuidance, "api_deployment.access_guidance", "", "Text briefly describing how to access the...")

	CreateApiDeploymentCmd.Flags().StringArrayVar(&CreateApiDeploymentInputApiDeploymentLabels, "api_deployment.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	CreateApiDeploymentCmd.Flags().StringArrayVar(&CreateApiDeploymentInputApiDeploymentAnnotations, "api_deployment.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentInput.ApiDeploymentId, "api_deployment_id", "", "Required. The ID to use for the deployment, which...")

	CreateApiDeploymentCmd.Flags().StringVar(&CreateApiDeploymentFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var CreateApiDeploymentCmd = &cobra.Command{
	Use:   "create-api-deployment",
	Short: "CreateApiDeployment creates a specified...",
	Long:  "CreateApiDeployment creates a specified deployment.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if CreateApiDeploymentFromFile == "" {

			cmd.MarkFlagRequired("parent")

			cmd.MarkFlagRequired("api_deployment_id")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if CreateApiDeploymentFromFile != "" {
			in, err = os.Open(CreateApiDeploymentFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &CreateApiDeploymentInput)
			if err != nil {
				return err
			}

		}

		if len(CreateApiDeploymentInputApiDeploymentLabels) > 0 {
			CreateApiDeploymentInput.ApiDeployment.Labels = make(map[string]string)
		}
		for _, item := range CreateApiDeploymentInputApiDeploymentLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateApiDeploymentInput.ApiDeployment.Labels[split[0]] = split[1]
		}

		if len(CreateApiDeploymentInputApiDeploymentAnnotations) > 0 {
			CreateApiDeploymentInput.ApiDeployment.Annotations = make(map[string]string)
		}
		for _, item := range CreateApiDeploymentInputApiDeploymentAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateApiDeploymentInput.ApiDeployment.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "CreateApiDeployment", &CreateApiDeploymentInput)
		}
		resp, err := RegistryClient.CreateApiDeployment(ctx, &CreateApiDeploymentInput)
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
