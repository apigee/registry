package cmd

import (
	"context"
	"fmt"
	"log"

	"apigov.dev/registry/cmd/registry/connection"
	"apigov.dev/registry/gapic"
	"apigov.dev/registry/models"
	rpc "apigov.dev/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/status"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources in the API model.",
	Long:  "List resources in the API model.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := connection.NewClient()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ctx := context.TODO()

		name := args[0]
		if m := models.ProjectsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listProjects(ctx, client, m[0])
		} else if m := models.ProductsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listProducts(ctx, client, m[0])
		} else if m := models.VersionsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listVersions(ctx, client, m[0])
		} else if m := models.SpecsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listSpecs(ctx, client, m[0], func(spec *rpc.Spec) error {
				fmt.Println(spec.Name)
				return nil
			})
		} else if m := models.PropertiesRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listProperties(ctx, client, m[0])

		} else if m := models.ProjectRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listProjects(ctx, client, segments)
			} else {
				err = getProject(ctx, client, segments)
			}
		} else if m := models.ProductRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listProducts(ctx, client, segments)
			} else {
				err = getProduct(ctx, client, segments)
			}
		} else if m := models.VersionRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listVersions(ctx, client, segments)
			} else {
				err = getVersion(ctx, client, segments)
			}
		} else if m := models.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listSpecs(ctx, client, segments, func(spec *rpc.Spec) error {
					fmt.Println(spec.Name)
					return nil
				})
			} else {
				_, err = getSpec(ctx, client, segments)
			}
		} else if m := models.PropertyRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listProperties(ctx, client, segments)
			} else {
				err = getProperty(ctx, client, segments)
			}
		} else {
			fmt.Printf("unsupported argument(s): %+v\n", args)
		}
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				log.Printf("%s", err.Error())
			} else {
				log.Printf("%s", st.Message())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listProjects(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {
	request := &rpc.ListProjectsRequest{}
	it := client.ListProjects(ctx, request)
	for {
		project, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		fmt.Println(project.Name)
	}
	return nil
}

func listProducts(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {
	request := &rpc.ListProductsRequest{
		Parent: "projects/" + segments[1],
	}
	filter := ""
	if len(segments) == 4 {
		filter = "product_id == '" + segments[3] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListProducts(ctx, request)
	for {
		product, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		fmt.Println(product.Name)
	}
	return nil
}

func listVersions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {
	request := &rpc.ListVersionsRequest{
		Parent: "projects/" + segments[1] + "/products/" + segments[2],
	}
	filter := ""
	if len(segments) == 4 {
		filter = "version_id == '" + segments[3] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListVersions(ctx, request)
	for {
		version, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		fmt.Println(version.Name)
	}
	return nil
}

type specHandler func(*rpc.Spec) error

func listSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler specHandler) error {
	request := &rpc.ListSpecsRequest{
		Parent: "projects/" + segments[1] + "/products/" + segments[2] + "/versions/" + segments[3],
	}
	filter := ""
	if len(segments) == 5 {
		filter = "spec_id == '" + segments[4] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListSpecs(ctx, request)
	for {
		spec, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(spec)
	}
	return nil
}

func listProperties(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {
	request := &rpc.ListPropertiesRequest{
		Parent: "projects/" + segments[1],
	}
	it := client.ListProperties(ctx, request)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		fmt.Println(property.Name)
	}
	return nil
}

func getProject(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {
	request := &rpc.GetProjectRequest{
		Name: "projects/" + segments[1],
	}
	project, err := client.GetProject(ctx, request)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", project)
	return nil
}

func getProduct(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {
	request := &rpc.GetProductRequest{
		Name: "projects/" + segments[1] + "/products/" + segments[2],
	}
	product, err := client.GetProduct(ctx, request)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", product)
	return nil
}

func getVersion(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {
	request := &rpc.GetVersionRequest{
		Name: "projects/" + segments[1] + "/products/" + segments[2] + "/versions/" + segments[3],
	}
	version, err := client.GetVersion(ctx, request)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", version)
	return nil
}

func getSpec(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) (*rpc.Spec, error) {
	request := &rpc.GetSpecRequest{
		Name: "projects/" + segments[1] + "/products/" + segments[2] + "/versions/" + segments[3] + "/specs/" + segments[4],
	}
	spec, err := client.GetSpec(ctx, request)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", spec)
	return spec, nil
}

func getProperty(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {
	request := &rpc.GetPropertyRequest{
		Name: "projects/" + segments[1] + "/properties/" + segments[2],
	}
	log.Printf("request %+v", request)
	property, err := client.GetProperty(ctx, request)
	if err != nil {
		log.Printf("%+s", err.Error())
	}
	fmt.Printf("%+v\n", property)
	print_property(property)
	return nil
}

func sliceContainsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
