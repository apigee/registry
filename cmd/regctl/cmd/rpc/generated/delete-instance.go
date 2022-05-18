// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteInstanceInput rpcpb.DeleteInstanceRequest

var DeleteInstanceFromFile string

var DeleteInstanceFollow bool

var DeleteInstancePollOperation string

func init() {
	ProvisioningServiceCmd.AddCommand(DeleteInstanceCmd)

	DeleteInstanceCmd.Flags().StringVar(&DeleteInstanceInput.Name, "name", "", "Required. The name of the Instance to delete. ...")

	DeleteInstanceCmd.Flags().StringVar(&DeleteInstanceFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

	DeleteInstanceCmd.Flags().BoolVar(&DeleteInstanceFollow, "follow", false, "Block until the long running operation completes")

	ProvisioningServiceCmd.AddCommand(DeleteInstancePollCmd)

	DeleteInstancePollCmd.Flags().BoolVar(&DeleteInstanceFollow, "follow", false, "Block until the long running operation completes")

	DeleteInstancePollCmd.Flags().StringVar(&DeleteInstancePollOperation, "operation", "", "Required. Operation name to poll for")

	DeleteInstancePollCmd.MarkFlagRequired("operation")

}

var DeleteInstanceCmd = &cobra.Command{
	Use:   "delete-instance",
	Short: "Deletes the Registry instance.",
	Long:  "Deletes the Registry instance.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteInstanceFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteInstanceFromFile != "" {
			in, err = os.Open(DeleteInstanceFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteInstanceInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Provisioning", "DeleteInstance", &DeleteInstanceInput)
		}
		resp, err := ProvisioningClient.DeleteInstance(ctx, &DeleteInstanceInput)
		if err != nil {
			return err
		}

		if !DeleteInstanceFollow {
			var s interface{}
			s = resp.Name()

			if OutputJSON {
				d := make(map[string]string)
				d["operation"] = resp.Name()
				s = d
			}

			printMessage(s)
			return err
		}

		err = resp.Wait(ctx)

		return err
	},
}

var DeleteInstancePollCmd = &cobra.Command{
	Use:   "poll-delete-instance",
	Short: "Poll the status of a DeleteInstanceOperation by name",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		op := ProvisioningClient.DeleteInstanceOperation(DeleteInstancePollOperation)

		if DeleteInstanceFollow {
			return op.Wait(ctx)
		}

		err = op.Poll(ctx)
		if err != nil {
			return err
		}

		if op.Done() {
			fmt.Println(fmt.Sprintf("Operation %s is done", op.Name()))
		} else {
			fmt.Println(fmt.Sprintf("Operation %s not done", op.Name()))
		}

		return err
	},
}
