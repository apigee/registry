// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetArtifactInput rpcpb.GetArtifactRequest

var GetArtifactFromFile string

func init() {
	RegistryServiceCmd.AddCommand(GetArtifactCmd)

	GetArtifactCmd.Flags().StringVar(&GetArtifactInput.Name, "name", "", "Required. The name of the artifact to retrieve. ...")

	GetArtifactCmd.Flags().StringVar(&GetArtifactFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetArtifactCmd = &cobra.Command{
	Use:   "get-artifact",
	Short: "GetArtifact returns a specified artifact.",
	Long:  "GetArtifact returns a specified artifact.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetArtifactFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetArtifactFromFile != "" {
			in, err = os.Open(GetArtifactFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetArtifactInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "GetArtifact", &GetArtifactInput)
		}
		resp, err := RegistryClient.GetArtifact(ctx, &GetArtifactInput)
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
