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

// productsCmd represents the products command
var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Returns a list of products.",
	Long:  "Returns a list of products.",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := connection.NewClient()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ctx := context.TODO()
		flagset := cmd.LocalFlags()
		projectID, err := flagset.GetString("project")

		request := &rpc.ListProductsRequest{}
		request.Parent = "projects/" + projectID
		it := client.ListProducts(ctx, request)
		fmt.Println("The names of your products:")
		for {
			product, err := it.Next()
			if err == iterator.Done {
				break
			} else if err != nil {
				log.Fatalf("%s", err.Error())
			}
			fmt.Println(product.Name)
		}
	},
}

func init() {
	listCmd.AddCommand(productsCmd)
	productsCmd.Flags().String("project", "", "Project identifier")
}
