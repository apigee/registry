package cmd

import (
	"context"
	"fmt"
	"log"

	"apigov.dev/flame/cmd/flame/connection"
	rpc "apigov.dev/flame/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

// versionsCmd represents the versions command
var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Returns a list of versions.",
	Long:  "Returns a list of versions.",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := connection.NewClient()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ctx := context.TODO()
		flagset := cmd.LocalFlags()
		projectID, err := flagset.GetString("project")
		productID, err := flagset.GetString("product")

		request := &rpc.ListVersionsRequest{}
		request.Parent = "projects/" + projectID + "/products/" + productID
		it := client.ListVersions(ctx, request)
		fmt.Println("The names of your versions:")
		for {
			version, err := it.Next()
			if err == iterator.Done {
				break
			} else if err != nil {
				log.Fatalf("%s", err.Error())
			}
			fmt.Println(version.Name)
		}
	},
}

func init() {
	listCmd.AddCommand(versionsCmd)
	versionsCmd.Flags().String("project", "", "Project identifier")
	versionsCmd.Flags().String("product", "", "Product identifier")
}
