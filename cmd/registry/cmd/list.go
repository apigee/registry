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

type projectHandler func(*rpc.Project)
type productHandler func(*rpc.Product)
type versionHandler func(*rpc.Version)
type specHandler func(*rpc.Spec)
type propertyHandler func(*rpc.Property)
type labelHandler func(*rpc.Label)

func printProject(project *rpc.Project) {
	fmt.Println(project.Name)
}

func printProduct(product *rpc.Product) {
	fmt.Println(product.Name)
}

func printVersion(version *rpc.Version) {
	fmt.Println(version.Name)
}

func printSpec(spec *rpc.Spec) {
	fmt.Println(spec.Name)
}

func printProperty(property *rpc.Property) {
	fmt.Println(property.Name)
}

func printLabel(label *rpc.Label) {
	fmt.Println(label.Name)
}

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
			err = listProjects(ctx, client, m[0], printProject)
		} else if m := models.ProductsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listProducts(ctx, client, m[0], printProduct)
		} else if m := models.VersionsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listVersions(ctx, client, m[0], printVersion)
		} else if m := models.SpecsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listSpecs(ctx, client, m[0], printSpec)
		} else if m := models.PropertiesRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listProperties(ctx, client, m[0], printProperty)
		} else if m := models.LabelsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listLabels(ctx, client, m[0], printLabel)

		} else if m := models.ProjectRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listProjects(ctx, client, segments, printProject)
			} else {
				err = getProject(ctx, client, segments, printProject)
			}
		} else if m := models.ProductRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listProducts(ctx, client, segments, printProduct)
			} else {
				err = getProduct(ctx, client, segments, printProduct)
			}
		} else if m := models.VersionRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listVersions(ctx, client, segments, printVersion)
			} else {
				err = getVersion(ctx, client, segments, printVersion)
			}
		} else if m := models.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listSpecs(ctx, client, segments, printSpec)
			} else {
				_, err = getSpec(ctx, client, segments, printSpec)
			}
		} else if m := models.PropertyRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listProperties(ctx, client, segments, printProperty)
			} else {
				err = getProperty(ctx, client, segments, printProperty)
			}
		} else if m := models.LabelRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listLabels(ctx, client, segments, printLabel)
			} else {
				err = getLabel(ctx, client, segments, printLabel)
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

var filterFlag string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&filterFlag, "filter", "", "Filter option to send with list calls")
}

func listProjects(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler projectHandler) error {
	request := &rpc.ListProjectsRequest{}
	filter := filterFlag
	if len(segments) == 2 && segments[1] != "-" {
		filter = "project_id == '" + segments[1] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListProjects(ctx, request)
	for {
		project, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(project)
	}
	return nil
}

func listProducts(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler productHandler) error {
	request := &rpc.ListProductsRequest{
		Parent: "projects/" + segments[1],
	}
	filter := filterFlag
	if len(segments) == 3 && segments[2] != "-" {
		filter = "product_id == '" + segments[2] + "'"
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
		handler(product)
	}
	return nil
}

func listVersions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler versionHandler) error {
	request := &rpc.ListVersionsRequest{
		Parent: "projects/" + segments[1] + "/products/" + segments[2],
	}
	filter := filterFlag
	if len(segments) == 4 && segments[3] != "-" {
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
		handler(version)
	}
	return nil
}

func listSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler specHandler) error {
	request := &rpc.ListSpecsRequest{
		Parent: "projects/" + segments[1] + "/products/" + segments[2] + "/versions/" + segments[3],
	}
	filter := filterFlag
	if len(segments) == 5 && segments[4] != "-" {
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
	segments []string,
	handler propertyHandler) error {
	request := &rpc.ListPropertiesRequest{
		Parent: "projects/" + segments[1],
	}
	filter := filterFlag
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListProperties(ctx, request)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(property)
	}
	return nil
}

func listLabels(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler labelHandler) error {
	request := &rpc.ListLabelsRequest{
		Parent: "projects/" + segments[1],
	}
	filter := filterFlag
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListLabels(ctx, request)
	for {
		label, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(label)
	}
	return nil
}

func getProject(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler projectHandler) error {
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
	segments []string,
	handler productHandler) error {
	request := &rpc.GetProductRequest{
		Name: "projects/" + segments[1] + "/products/" + segments[2],
	}
	product, err := client.GetProduct(ctx, request)
	if err != nil {
		return err
	}
	handler(product)
	return nil
}

func getVersion(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler versionHandler) error {
	request := &rpc.GetVersionRequest{
		Name: "projects/" + segments[1] + "/products/" + segments[2] + "/versions/" + segments[3],
	}
	version, err := client.GetVersion(ctx, request)
	if err != nil {
		return err
	}
	handler(version)
	return nil
}

func getSpec(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler specHandler) (*rpc.Spec, error) {
	request := &rpc.GetSpecRequest{
		Name: "projects/" + segments[1] + "/products/" + segments[2] + "/versions/" + segments[3] + "/specs/" + segments[4],
	}
	spec, err := client.GetSpec(ctx, request)
	if err != nil {
		return nil, err
	}
	handler(spec)
	return spec, nil
}

func getProperty(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler propertyHandler) error {
	request := &rpc.GetPropertyRequest{
		Name: "projects/" + segments[1] + "/properties/" + segments[2],
	}
	log.Printf("request %+v", request)
	property, err := client.GetProperty(ctx, request)
	if err != nil {
		log.Printf("%+s", err.Error())
	}
	handler(property)
	fmt.Printf("%+v\n", property)
	print_property(property)
	return nil
}

func getLabel(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler labelHandler) error {
	request := &rpc.GetLabelRequest{
		Name: "projects/" + segments[1] + "/labels/" + segments[2],
	}
	log.Printf("request %+v", request)
	label, err := client.GetLabel(ctx, request)
	if err != nil {
		log.Printf("%+s", err.Error())
	}
	fmt.Printf("%+v\n", label)
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
