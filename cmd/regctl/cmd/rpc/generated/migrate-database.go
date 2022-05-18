// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var MigrateDatabaseInput rpcpb.MigrateDatabaseRequest

var MigrateDatabaseFromFile string

var MigrateDatabaseFollow bool

var MigrateDatabasePollOperation string

func init() {
	AdminServiceCmd.AddCommand(MigrateDatabaseCmd)

	MigrateDatabaseCmd.Flags().StringVar(&MigrateDatabaseInput.Kind, "kind", "", "A string describing the kind of migration to...")

	MigrateDatabaseCmd.Flags().StringVar(&MigrateDatabaseFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

	MigrateDatabaseCmd.Flags().BoolVar(&MigrateDatabaseFollow, "follow", false, "Block until the long running operation completes")

	AdminServiceCmd.AddCommand(MigrateDatabasePollCmd)

	MigrateDatabasePollCmd.Flags().BoolVar(&MigrateDatabaseFollow, "follow", false, "Block until the long running operation completes")

	MigrateDatabasePollCmd.Flags().StringVar(&MigrateDatabasePollOperation, "operation", "", "Required. Operation name to poll for")

	MigrateDatabasePollCmd.MarkFlagRequired("operation")

}

var MigrateDatabaseCmd = &cobra.Command{
	Use:   "migrate-database",
	Short: "MigrateDatabase attempts to migrate the database...",
	Long:  "MigrateDatabase attempts to migrate the database to the current schema.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if MigrateDatabaseFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if MigrateDatabaseFromFile != "" {
			in, err = os.Open(MigrateDatabaseFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &MigrateDatabaseInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Admin", "MigrateDatabase", &MigrateDatabaseInput)
		}
		resp, err := AdminClient.MigrateDatabase(ctx, &MigrateDatabaseInput)
		if err != nil {
			return err
		}

		if !MigrateDatabaseFollow {
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

var MigrateDatabasePollCmd = &cobra.Command{
	Use:   "poll-migrate-database",
	Short: "Poll the status of a MigrateDatabaseOperation by name",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		op := AdminClient.MigrateDatabaseOperation(MigrateDatabasePollOperation)

		if MigrateDatabaseFollow {
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
