// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetArtifactContentsInput rpcpb.GetArtifactContentsRequest

var GetArtifactContentsFromFile string

func init() {
	RegistryServiceCmd.AddCommand(GetArtifactContentsCmd)

	GetArtifactContentsCmd.Flags().StringVar(&GetArtifactContentsInput.Name, "name", "", "Required. The name of the artifact whose contents...")

	GetArtifactContentsCmd.Flags().StringVar(&GetArtifactContentsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetArtifactContentsCmd = &cobra.Command{
	Use:   "get-artifact-contents",
	Short: "GetArtifactContents returns the contents of a...",
	Long:  "GetArtifactContents returns the contents of a specified artifact.  If artifacts are stored with GZip compression, the default behavior  is to return...",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetArtifactContentsFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetArtifactContentsFromFile != "" {
			in, err = os.Open(GetArtifactContentsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetArtifactContentsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "GetArtifactContents", &GetArtifactContentsInput)
		}
		resp, err := RegistryClient.GetArtifactContents(ctx, &GetArtifactContentsInput)
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
