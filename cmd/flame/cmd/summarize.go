package cmd

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"apigov.dev/flame/cmd/flame/connection"
	"apigov.dev/flame/gapic"
	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"github.com/googleapis/gnostic/compiler"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// summarizeCmd represents the summarize command
var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Generate a summary of an API spec",
	Long:  `Generate a summary of an API spec.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("summarize called %+v", args)
		client, err := connection.NewClient()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ctx := context.TODO()
		name := args[0]
		if m := models.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				// iterate through a collection of specs and summarize each
				completions := make(chan int)
				processes := 0
				err = listSpecs(ctx, client, segments, func(spec *rpc.Spec) error {
					fmt.Println(spec.Name)
					m := models.SpecRegexp().FindAllStringSubmatch(spec.Name, -1)
					if m != nil {
						processes++
						go func() {
							summarizeSpec(ctx, client, m[0])
							completions <- 1
						}()
						return nil
					}
					return nil
				})
				for i := 0; i < processes; i++ {
					<-completions
				}

			} else {
				err := summarizeSpec(ctx, client, segments)
				if err != nil {
					log.Printf("%s", err.Error())
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}

func resourceNameOfSpec(segments []string) string {
	if len(segments) == 4 {
		return "projects/" + segments[0] +
			"/products/" + segments[1] +
			"/versions/" + segments[2] +
			"/specs/" + segments[3]
	}
	return ""
}

func summarizeSpec(ctx context.Context,
	client *gapic.FlameClient,
	segments []string) error {

	name := resourceNameOfSpec(segments[1:])

	request := &rpc.GetSpecRequest{
		Name: name,
		View: rpc.SpecView_FULL,
	}
	spec, err := client.GetSpec(ctx, request)
	if err != nil {
		return err
	}

	if strings.HasPrefix(spec.GetStyle(), "openapi/v2") {
		var data []byte
		if strings.Contains(spec.GetStyle(), "+gzip") {
			gr, err := gzip.NewReader(bytes.NewBuffer(spec.GetContents()))
			defer gr.Close()
			data, err = ioutil.ReadAll(gr)
			if err != nil {
				return err
			}
		} else {
			data = spec.GetContents()
		}
		info, err := compiler.ReadInfoFromBytes(spec.GetName(), data)
		if err != nil {
			return err
		}
		document, err := openapi_v2.NewDocument(info, compiler.NewContextWithExtensions("$root", nil, nil))
		if err != nil {
			return err
		}
		summary := summarizeOpenAPIv2Document(document)
		bytes, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", string(bytes))

		projectID := segments[1]

		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "summary"
		property.Value = &rpc.Property_StringValue{StringValue: string(bytes)}
		err = setProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}

		property.Subject = spec.GetName()
		property.Relation = "operationCount"
		operationCount := summary.GetCount + summary.PostCount + summary.PutCount + summary.DeleteCount
		property.Value = &rpc.Property_Int64Value{Int64Value: int64(operationCount)}
		err = setProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}

		property.Subject = spec.GetName()
		property.Relation = "schemaCount"
		property.Value = &rpc.Property_Int64Value{Int64Value: int64(summary.SchemaCount)}
		err = setProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}
	}
	return nil
}

func hash(name string) string {
	h := sha1.New()
	io.WriteString(h, name)
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func setProperty(ctx context.Context, client *gapic.FlameClient, projectID string, property *rpc.Property) error {
	propertyID := hash(property.Subject + "/" + property.Relation)
	property.Name = "projects/" + projectID + "/properties/" + propertyID

	request := &rpc.CreatePropertyRequest{}
	request.Property = property
	request.Parent = "projects/" + projectID
	request.PropertyId = propertyID
	newProperty, err := client.CreateProperty(ctx, request)
	if err != nil {
		code := status.Code(err)
		if code == codes.AlreadyExists {
			fmt.Printf("already exists\n")
			request := &rpc.UpdatePropertyRequest{}
			request.Property = property
			updatedProperty, err := client.UpdateProperty(ctx, request)
			if err != nil {
				return err
			}
			fmt.Printf("updated %+v\n", updatedProperty)
		} else {
			return err
		}
	} else {
		fmt.Printf("created %+v\n", newProperty)
	}
	return nil
}

// Summary ...
type Summary struct {
	Title            string
	Description      string
	Version          string
	SchemaCount      int
	PathCount        int
	GetCount         int
	PostCount        int
	PutCount         int
	DeleteCount      int
	TagCount         int
	VendorExtensions []string
}

// NewSummary ...
func NewSummary() *Summary {
	s := &Summary{}
	s.VendorExtensions = make([]string, 0)
	return s
}

func (s *Summary) addVendorExtension(name string) {
	for _, n := range s.VendorExtensions {
		if n == name {
			return
		}
	}
	s.VendorExtensions = append(s.VendorExtensions, name)
}

func truncateString(str string, num int) string {
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		return str[0:num] + "..."
	}
	return str
}

func summarizeOpenAPIv2Document(document *openapi_v2.Document) *Summary {
	summary := NewSummary()

	if document.Info != nil {
		summary.Title = document.Info.Title
		summary.Description = truncateString(document.Info.Description, 240)
		summary.Version = document.Info.Version
	}

	if document.Definitions != nil && document.Definitions.AdditionalProperties != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			summarizeSchema(summary, pair.Value)
		}
	}

	for _, pair := range document.Paths.Path {
		summary.PathCount++
		v := pair.Value
		if v.Get != nil {
			summary.GetCount++
			summarizeOperation(summary, v.Get)
		}
		if v.Post != nil {
			summary.PostCount++
			summarizeOperation(summary, v.Post)
		}
		if v.Put != nil {
			summary.PutCount++
			summarizeOperation(summary, v.Put)
		}
		if v.Delete != nil {
			summary.DeleteCount++
			summarizeOperation(summary, v.Delete)
		}
	}
	for _, tag := range document.Tags {
		summary.TagCount++
		summarizeVendorExtension(summary, tag.VendorExtension)
	}
	return summary
}

func summarizeOperation(summary *Summary, operation *openapi_v2.Operation) {
	summarizeVendorExtension(summary, operation.Responses.VendorExtension)
	summarizeVendorExtension(summary, operation.VendorExtension)
}

func summarizeSchema(summary *Summary, schema *openapi_v2.Schema) {
	summary.SchemaCount++
	if schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			summarizeSchema(summary, pair.Value)
		}
	}
	summarizeVendorExtension(summary, schema.VendorExtension)
}

func summarizeVendorExtension(summary *Summary, vendorExtension []*openapi_v2.NamedAny) {
	if len(vendorExtension) > 0 {
		for _, v := range vendorExtension {
			summary.addVendorExtension(v.Name)
		}
	}
}
