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

var CreateApiVersionInput rpcpb.CreateApiVersionRequest

var CreateApiVersionFromFile string

var CreateApiVersionInputApiVersionLabels []string

var CreateApiVersionInputApiVersionAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(CreateApiVersionCmd)

	CreateApiVersionInput.ApiVersion = new(rpcpb.ApiVersion)

	CreateApiVersionCmd.Flags().StringVar(&CreateApiVersionInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	CreateApiVersionCmd.Flags().StringVar(&CreateApiVersionInput.ApiVersion.Name, "api_version.name", "", "Resource name.")

	CreateApiVersionCmd.Flags().StringVar(&CreateApiVersionInput.ApiVersion.DisplayName, "api_version.display_name", "", "Human-meaningful name.")

	CreateApiVersionCmd.Flags().StringVar(&CreateApiVersionInput.ApiVersion.Description, "api_version.description", "", "A detailed description.")

	CreateApiVersionCmd.Flags().StringVar(&CreateApiVersionInput.ApiVersion.State, "api_version.state", "", "A user-definable description of the lifecycle...")

	CreateApiVersionCmd.Flags().StringArrayVar(&CreateApiVersionInputApiVersionLabels, "api_version.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	CreateApiVersionCmd.Flags().StringArrayVar(&CreateApiVersionInputApiVersionAnnotations, "api_version.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	CreateApiVersionCmd.Flags().StringVar(&CreateApiVersionInput.ApiVersion.PrimarySpec, "api_version.primary_spec", "", "The primary spec for this version.  Format:...")

	CreateApiVersionCmd.Flags().StringVar(&CreateApiVersionInput.ApiVersionId, "api_version_id", "", "Required. The ID to use for the version, which...")

	CreateApiVersionCmd.Flags().StringVar(&CreateApiVersionFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var CreateApiVersionCmd = &cobra.Command{
	Use:   "create-api-version",
	Short: "CreateApiVersion creates a specified version.",
	Long:  "CreateApiVersion creates a specified version.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if CreateApiVersionFromFile == "" {

			cmd.MarkFlagRequired("parent")

			cmd.MarkFlagRequired("api_version_id")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if CreateApiVersionFromFile != "" {
			in, err = os.Open(CreateApiVersionFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &CreateApiVersionInput)
			if err != nil {
				return err
			}

		}

		if len(CreateApiVersionInputApiVersionLabels) > 0 {
			CreateApiVersionInput.ApiVersion.Labels = make(map[string]string)
		}
		for _, item := range CreateApiVersionInputApiVersionLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateApiVersionInput.ApiVersion.Labels[split[0]] = split[1]
		}

		if len(CreateApiVersionInputApiVersionAnnotations) > 0 {
			CreateApiVersionInput.ApiVersion.Annotations = make(map[string]string)
		}
		for _, item := range CreateApiVersionInputApiVersionAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateApiVersionInput.ApiVersion.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "CreateApiVersion", &CreateApiVersionInput)
		}
		resp, err := RegistryClient.CreateApiVersion(ctx, &CreateApiVersionInput)
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
