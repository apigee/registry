// Code generated. DO NOT EDIT.

package main

import (
	"github.com/spf13/cobra"

	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"fmt"
)

var GetStatusInput emptypb.Empty

func init() {
	AdminServiceCmd.AddCommand(GetStatusCmd)

}

var GetStatusCmd = &cobra.Command{
	Use:   "get-status",
	Short: "GetStatus returns the status of the service. ...",
	Long:  "GetStatus returns the status of the service.  GetStatus is for verifying open source deployments only  and is not included in hosted versions of the...",
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		if Verbose {
			printVerboseInput("Admin", "GetStatus", &GetStatusInput)
		}
		resp, err := AdminClient.GetStatus(ctx, &GetStatusInput)

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(resp)

		return err
	},
}
