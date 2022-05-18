// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var TagApiSpecRevisionInput rpcpb.TagApiSpecRevisionRequest

var TagApiSpecRevisionFromFile string

func init() {
	RegistryServiceCmd.AddCommand(TagApiSpecRevisionCmd)

	TagApiSpecRevisionCmd.Flags().StringVar(&TagApiSpecRevisionInput.Name, "name", "", "Required. The name of the spec to be tagged,...")

	TagApiSpecRevisionCmd.Flags().StringVar(&TagApiSpecRevisionInput.Tag, "tag", "", "Required. The tag to apply.  The tag should be at...")

	TagApiSpecRevisionCmd.Flags().StringVar(&TagApiSpecRevisionFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var TagApiSpecRevisionCmd = &cobra.Command{
	Use:   "tag-api-spec-revision",
	Short: "TagApiSpecRevision adds a tag to a specified...",
	Long:  "TagApiSpecRevision adds a tag to a specified revision of a spec.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if TagApiSpecRevisionFromFile == "" {

			cmd.MarkFlagRequired("name")

			cmd.MarkFlagRequired("tag")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if TagApiSpecRevisionFromFile != "" {
			in, err = os.Open(TagApiSpecRevisionFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &TagApiSpecRevisionInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "TagApiSpecRevision", &TagApiSpecRevisionInput)
		}
		resp, err := RegistryClient.TagApiSpecRevision(ctx, &TagApiSpecRevisionInput)
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
