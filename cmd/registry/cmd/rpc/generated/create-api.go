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

var CreateApiInput rpcpb.CreateApiRequest

var CreateApiFromFile string

var CreateApiInputApiLabels []string

var CreateApiInputApiAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(CreateApiCmd)

	CreateApiInput.Api = new(rpcpb.Api)

	CreateApiCmd.Flags().StringVar(&CreateApiInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	CreateApiCmd.Flags().StringVar(&CreateApiInput.Api.Name, "api.name", "", "Resource name.")

	CreateApiCmd.Flags().StringVar(&CreateApiInput.Api.DisplayName, "api.display_name", "", "Human-meaningful name.")

	CreateApiCmd.Flags().StringVar(&CreateApiInput.Api.Description, "api.description", "", "A detailed description.")

	CreateApiCmd.Flags().StringVar(&CreateApiInput.Api.Availability, "api.availability", "", "A user-definable description of the availability...")

	CreateApiCmd.Flags().StringVar(&CreateApiInput.Api.RecommendedVersion, "api.recommended_version", "", "The recommended version of the API.  Format:...")

	CreateApiCmd.Flags().StringVar(&CreateApiInput.Api.RecommendedDeployment, "api.recommended_deployment", "", "The recommended deployment of the API.  Format:...")

	CreateApiCmd.Flags().StringArrayVar(&CreateApiInputApiLabels, "api.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	CreateApiCmd.Flags().StringArrayVar(&CreateApiInputApiAnnotations, "api.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	CreateApiCmd.Flags().StringVar(&CreateApiInput.ApiId, "api_id", "", "Required. The ID to use for the api, which will...")

	CreateApiCmd.Flags().StringVar(&CreateApiFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var CreateApiCmd = &cobra.Command{
	Use:   "create-api",
	Short: "CreateApi creates a specified API.",
	Long:  "CreateApi creates a specified API.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if CreateApiFromFile == "" {

			cmd.MarkFlagRequired("parent")

			cmd.MarkFlagRequired("api_id")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if CreateApiFromFile != "" {
			in, err = os.Open(CreateApiFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &CreateApiInput)
			if err != nil {
				return err
			}

		}

		if len(CreateApiInputApiLabels) > 0 {
			CreateApiInput.Api.Labels = make(map[string]string)
		}
		for _, item := range CreateApiInputApiLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateApiInput.Api.Labels[split[0]] = split[1]
		}

		if len(CreateApiInputApiAnnotations) > 0 {
			CreateApiInput.Api.Annotations = make(map[string]string)
		}
		for _, item := range CreateApiInputApiAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateApiInput.Api.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "CreateApi", &CreateApiInput)
		}
		resp, err := RegistryClient.CreateApi(ctx, &CreateApiInput)
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
