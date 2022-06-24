// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var CreateInstanceInput rpcpb.CreateInstanceRequest

var CreateInstanceFromFile string

var CreateInstanceFollow bool

var CreateInstancePollOperation string

func init() {
	ProvisioningServiceCmd.AddCommand(CreateInstanceCmd)

	CreateInstanceInput.Instance = new(rpcpb.Instance)

	CreateInstanceInput.Instance.Config = new(rpcpb.Instance_Config)

	CreateInstanceCmd.Flags().StringVar(&CreateInstanceInput.Parent, "parent", "", "Required. Parent resource of the Instance, of the...")

	CreateInstanceCmd.Flags().StringVar(&CreateInstanceInput.InstanceId, "instance_id", "", "Required. Identifier to assign to the Instance....")

	CreateInstanceCmd.Flags().StringVar(&CreateInstanceInput.Instance.Name, "instance.name", "", "Format: `projects/*/locations/*/instance`. ...")

	CreateInstanceCmd.Flags().StringVar(&CreateInstanceInput.Instance.Config.CmekKeyName, "instance.config.cmek_key_name", "", "Required. The Customer Managed Encryption Key...")

	CreateInstanceCmd.Flags().StringVar(&CreateInstanceFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

	CreateInstanceCmd.Flags().BoolVar(&CreateInstanceFollow, "follow", false, "Block until the long running operation completes")

	ProvisioningServiceCmd.AddCommand(CreateInstancePollCmd)

	CreateInstancePollCmd.Flags().BoolVar(&CreateInstanceFollow, "follow", false, "Block until the long running operation completes")

	CreateInstancePollCmd.Flags().StringVar(&CreateInstancePollOperation, "operation", "", "Required. Operation name to poll for")

	CreateInstancePollCmd.MarkFlagRequired("operation")

}

var CreateInstanceCmd = &cobra.Command{
	Use:   "create-instance",
	Short: "Provisions instance resources for the Registry.",
	Long:  "Provisions instance resources for the Registry.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if CreateInstanceFromFile == "" {

			cmd.MarkFlagRequired("parent")

			cmd.MarkFlagRequired("instance_id")

			cmd.MarkFlagRequired("instance.config.cmek_key_name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if CreateInstanceFromFile != "" {
			in, err = os.Open(CreateInstanceFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &CreateInstanceInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Provisioning", "CreateInstance", &CreateInstanceInput)
		}
		resp, err := ProvisioningClient.CreateInstance(ctx, &CreateInstanceInput)
		if err != nil {
			return err
		}

		if !CreateInstanceFollow {
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

		result, err := resp.Wait(ctx)
		if err != nil {
			return err
		}

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(result)

		return err
	},
}

var CreateInstancePollCmd = &cobra.Command{
	Use:   "poll-create-instance",
	Short: "Poll the status of a CreateInstanceOperation by name",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		op := ProvisioningClient.CreateInstanceOperation(CreateInstancePollOperation)

		if CreateInstanceFollow {
			resp, err := op.Wait(ctx)
			if err != nil {
				return err
			}

			if Verbose {
				fmt.Print("Output: ")
			}
			printMessage(resp)
			return err
		}

		resp, err := op.Poll(ctx)
		if err != nil {
			return err
		} else if resp != nil {
			if Verbose {
				fmt.Print("Output: ")
			}

			printMessage(resp)
			return
		}

		if op.Done() {
			fmt.Println(fmt.Sprintf("Operation %s is done", op.Name()))
		} else {
			fmt.Println(fmt.Sprintf("Operation %s not done", op.Name()))
		}

		return err
	},
}
