package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"apigov.dev/registry/connection"
	rpcpb "apigov.dev/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const directory = "."
const project = "atlas"

var registryClient connection.Client

func notFound(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.NotFound
}

func alreadyExists(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.AlreadyExists
}

func main() {
	var err error

	ctx := context.Background()
	registryClient, err = connection.NewClient(ctx)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(-1)
	}
	completions := make(chan int)
	processes := 0

	// walk a directory hierarchy, uploading every API spec that matches a set of expected file names.
	err = filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, "swagger.yaml") || strings.HasSuffix(path, "swagger.json") {
				processes++
				go func() {
					err := handleSpec(path, "openapi/v2")
					if err != nil {
						fmt.Printf("%s\n", err.Error())
					}
					completions <- 1
				}()
			}
			if strings.HasSuffix(path, "openapi.yaml") || strings.HasSuffix(path, "openapi.json") {
				processes++
				go func() {
					err := handleSpec(path, "openapi/v3")
					if err != nil {
						fmt.Printf("%s\n", err.Error())
					}
					completions <- 1
				}()
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	for i := 0; i < processes; i++ {
		<-completions
		fmt.Printf("COMPLETE: %d\n", i+1)
	}
}

func handleSpec(path string, style string) error {
	// Compute the API name from the path to the spec file.
	name := strings.TrimPrefix(path, directory)
	parts := strings.Split(name, "/")
	spec := parts[len(parts)-1]
	version := parts[len(parts)-2]
	product := strings.Join(parts[0:len(parts)-2], "/")
	fmt.Printf("product:%+v version:%+v spec:%+v \n", product, version, spec)
	// Upload the spec for the specified product, version, and style
	return uploadSpec(product, version, style, path)
}

func uploadSpec(productName, version, style, path string) error {
	ctx := context.TODO()
	product := strings.Replace(productName, "/", "-", -1)
	// If the API product does not exist, create it.
	{
		request := &rpcpb.GetProductRequest{}
		request.Name = "projects/" + project + "/products/" + product
		_, err := registryClient.GetProduct(ctx, request)
		if notFound(err) {
			request := &rpcpb.CreateProductRequest{}
			request.Parent = "projects/" + project
			request.ProductId = product
			request.Product = &rpcpb.Product{}
			request.Product.DisplayName = productName
			response, err := registryClient.CreateProduct(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else if alreadyExists(err) {
				log.Printf("already exists %s/products/%s", request.Parent, request.ProductId)
			} else {
				log.Printf("failed to create %s/products/%s: %s",
					request.Parent, request.ProductId, err.Error())
			}
		} else if err != nil {
			return err
		}
	}
	// If the API version does not exist, create it.
	{
		request := &rpcpb.GetVersionRequest{}
		request.Name = "projects/" + project + "/products/" + product + "/versions/" + version
		_, err := registryClient.GetVersion(ctx, request)
		if notFound(err) {
			request := &rpcpb.CreateVersionRequest{}
			request.Parent = "projects/" + project + "/products/" + product
			request.VersionId = version
			request.Version = &rpcpb.Version{}
			response, err := registryClient.CreateVersion(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else if alreadyExists(err) {
				log.Printf("already exists %s/versions/%s", request.Parent, request.VersionId)
			} else {
				log.Printf("failed to create %s/versions/%s: %s",
					request.Parent, request.VersionId, err.Error())
			}
		} else if err != nil {
			return err
		}
	}
	// If the API spec does not exist, create it.
	{
		filename := filepath.Base(path)

		request := &rpcpb.GetSpecRequest{}
		request.Name = "projects/" + project + "/products/" + product +
			"/versions/" + version +
			"/specs/" + filename
		_, err := registryClient.GetSpec(ctx, request)
		if notFound(err) {
			fileBytes, err := ioutil.ReadFile(path)

			// gzip the spec before uploading it
			var buf bytes.Buffer
			zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
			_, err = zw.Write(fileBytes)
			if err != nil {
				log.Fatal(err)
			}
			if err := zw.Close(); err != nil {
				log.Fatal(err)
			}

			request := &rpcpb.CreateSpecRequest{}
			request.Parent = "projects/" + project + "/products/" + product +
				"/versions/" + version
			request.SpecId = filename
			request.Spec = &rpcpb.Spec{}
			request.Spec.Style = style + "+gzip"
			request.Spec.Filename = filename
			request.Spec.Contents = buf.Bytes()
			response, err := registryClient.CreateSpec(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else if alreadyExists(err) {
				log.Printf("already exists %s/specs/%s", request.Parent, request.SpecId)
			} else {
				details := fmt.Sprintf("contents-length: %d", len(request.Spec.Contents))
				log.Printf("failed to create %s/specs/%s: %s [%s]",
					request.Parent, request.SpecId, err.Error(), details)
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}
