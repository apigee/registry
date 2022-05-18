// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetApiSpecContentsInput rpcpb.GetApiSpecContentsRequest

var GetApiSpecContentsFromFile string

func init() {
	RegistryServiceCmd.AddCommand(GetApiSpecContentsCmd)

	GetApiSpecContentsCmd.Flags().StringVar(&GetApiSpecContentsInput.Name, "name", "", "Required. The name of the spec whose contents...")

	GetApiSpecContentsCmd.Flags().StringVar(&GetApiSpecContentsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetApiSpecContentsCmd = &cobra.Command{
	Use:   "get-api-spec-contents",
	Short: "GetApiSpecContents returns the contents of a...",
	Long:  "GetApiSpecContents returns the contents of a specified spec.  If specs are stored with GZip compression, the default behavior  is to return the spec...",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetApiSpecContentsFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetApiSpecContentsFromFile != "" {
			in, err = os.Open(GetApiSpecContentsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetApiSpecContentsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "GetApiSpecContents", &GetApiSpecContentsInput)
		}
		resp, err := RegistryClient.GetApiSpecContents(ctx, &GetApiSpecContentsInput)
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
