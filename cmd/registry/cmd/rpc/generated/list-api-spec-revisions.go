// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"google.golang.org/api/iterator"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var ListApiSpecRevisionsInput rpcpb.ListApiSpecRevisionsRequest

var ListApiSpecRevisionsFromFile string

func init() {
	RegistryServiceCmd.AddCommand(ListApiSpecRevisionsCmd)

	ListApiSpecRevisionsCmd.Flags().StringVar(&ListApiSpecRevisionsInput.Name, "name", "", "Required. The name of the spec to list revisions...")

	ListApiSpecRevisionsCmd.Flags().Int32Var(&ListApiSpecRevisionsInput.PageSize, "page_size", 10, "Default is 10. The maximum number of revisions to return per...")

	ListApiSpecRevisionsCmd.Flags().StringVar(&ListApiSpecRevisionsInput.PageToken, "page_token", "", "The page token, received from a previous...")

	ListApiSpecRevisionsCmd.Flags().StringVar(&ListApiSpecRevisionsInput.Filter, "filter", "", "An expression that can be used to filter the...")

	ListApiSpecRevisionsCmd.Flags().StringVar(&ListApiSpecRevisionsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ListApiSpecRevisionsCmd = &cobra.Command{
	Use:   "list-api-spec-revisions",
	Short: "ListApiSpecRevisions lists all revisions of a...",
	Long:  "ListApiSpecRevisions lists all revisions of a spec.  Revisions are returned in descending order of revision creation time.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ListApiSpecRevisionsFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ListApiSpecRevisionsFromFile != "" {
			in, err = os.Open(ListApiSpecRevisionsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ListApiSpecRevisionsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "ListApiSpecRevisions", &ListApiSpecRevisionsInput)
		}
		iter := RegistryClient.ListApiSpecRevisions(ctx, &ListApiSpecRevisionsInput)

		// populate iterator with a page
		_, err = iter.Next()
		if err != nil && err != iterator.Done {
			return err
		}

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(iter.Response)

		return err
	},
}
