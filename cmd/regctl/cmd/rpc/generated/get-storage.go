// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"fmt"
)

var GetStorageInput emptypb.Empty

func init() {
	AdminServiceCmd.AddCommand(GetStorageCmd)

}

var GetStorageCmd = &cobra.Command{
	Use:   "get-storage",
	Short: "GetStorage returns information about the storage...",
	Long:  "GetStorage returns information about the storage used by the service.  (-- api-linter: core::0131::request-message-name=disabled     ...",
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		if Verbose {
			printVerboseInput("Admin", "GetStorage", &GetStorageInput)
		}
		resp, err := AdminClient.GetStorage(ctx, &GetStorageInput)
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
