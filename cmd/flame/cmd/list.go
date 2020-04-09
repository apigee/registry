package cmd

import (
	"context"
	"fmt"
	"log"

	"apigov.dev/flame/cmd/flame/connection"
	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("list called with %+v\n", args)
		name := args[0]
		if m := models.ProductsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.ListProductsRequest{
				Parent: "projects/" + m[0][1],
			}
			it := client.ListProducts(ctx, request)
			for {
				product, err := it.Next()
				if err == iterator.Done {
					break
				} else if err != nil {
					log.Fatalf("%s", err.Error())
				}
				fmt.Println(product.Name)
			}
		} else if m := models.VersionsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.ListVersionsRequest{
				Parent: "projects/" + m[0][1] + "/products/" + m[0][2],
			}
			it := client.ListVersions(ctx, request)
			for {
				version, err := it.Next()
				if err == iterator.Done {
					break
				} else if err != nil {
					log.Fatalf("%s", err.Error())
				}
				fmt.Println(version.Name)
			}
		} else if m := models.SpecsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.ListSpecsRequest{
				Parent: "projects/" + m[0][1] + "/products/" + m[0][2] + "/versions/" + m[0][3],
			}
			it := client.ListSpecs(ctx, request)
			for {
				spec, err := it.Next()
				if err == iterator.Done {
					break
				} else if err != nil {
					log.Fatalf("%s", err.Error())
				}
				fmt.Println(spec.Name)
			}
		} else if m := models.FilesRegexp().FindAllStringSubmatch(name, -1); m != nil {
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.ListFilesRequest{
				Parent: "projects/" + m[0][1] + "/products/" + m[0][2] + "/versions/" + m[0][3] + "/specs/" + m[0][4],
			}
			it := client.ListFiles(ctx, request)
			for {
				file, err := it.Next()
				if err == iterator.Done {
					break
				} else if err != nil {
					log.Fatalf("%s", err.Error())
				}
				fmt.Println(file.Name)
			}
		} else if m := models.PropertiesRegexp().FindAllStringSubmatch(name, -1); m != nil {
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.ListPropertiesRequest{
				Parent: "projects/" + m[0][1],
			}
			it := client.ListProperties(ctx, request)
			for {
				property, err := it.Next()
				if err == iterator.Done {
					break
				} else if err != nil {
					log.Fatalf("%s", err.Error())
				}
				fmt.Println(property.Name)
			}
		} else if m := models.ProductRegexp().FindAllStringSubmatch(name, -1); m != nil {
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.GetProductRequest{
				Name: "projects/" + m[0][1] + "/products/" + m[0][2],
			}
			product, err := client.GetProduct(ctx, request)
			fmt.Printf("%+v\n", product)
		} else if m := models.VersionRegexp().FindAllStringSubmatch(name, -1); m != nil {
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.GetVersionRequest{
				Name: "projects/" + m[0][1] + "/products/" + m[0][2] + "/versions/" + m[0][3],
			}
			version, err := client.GetVersion(ctx, request)
			fmt.Printf("%+v\n", version)
		} else if m := models.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.GetSpecRequest{
				Name: "projects/" + m[0][1] + "/products/" + m[0][2] + "/versions/" + m[0][3] + "/specs/" + m[0][4],
			}
			spec, err := client.GetSpec(ctx, request)
			fmt.Printf("%+v\n", spec)
		} else if m := models.FileRegexp().FindAllStringSubmatch(name, -1); m != nil {
			fmt.Printf("FILE\n")
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.GetFileRequest{
				Name: "projects/" + m[0][1] + "/products/" + m[0][2] + "/versions/" + m[0][3] + "/specs/" + m[0][4] + "/files/" + m[0][5],
			}
			log.Printf("request %+v", request)
			file, err := client.GetFile(ctx, request)
			fmt.Printf("%+v\n", file)
		} else if m := models.PropertyRegexp().FindAllStringSubmatch(name, -1); m != nil {
			fmt.Printf("FILE\n")
			client, err := connection.NewClient()
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			ctx := context.TODO()
			request := &rpc.GetPropertyRequest{
				Name: "projects/" + m[0][1] + "/properties/" + m[0][2],
			}
			log.Printf("request %+v", request)
			property, err := client.GetProperty(ctx, request)
			fmt.Printf("%+v\n", property)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
