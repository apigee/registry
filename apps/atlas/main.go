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

	"apigov.dev/flame/client"
	"apigov.dev/flame/gapic"
	rpcpb "apigov.dev/flame/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const directory = "."

var flameClient *gapic.FlameClient

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

func main() {
	var err error

	flameClient, err = client.NewClient()
	completions := make(chan int)
	processes := 0

	err = filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, "swagger.yaml") {
				processes++
				go func() {
					handleSpec(path, "openapi-v2")
					completions <- 1
				}()
			}
			if strings.HasSuffix(path, "openapi.yaml") {
				processes++
				go func() {
					handleSpec(path, "openapi-v3")
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
	name := strings.TrimPrefix(path, directory)
	parts := strings.Split(name, "/")
	spec := parts[len(parts)-1]
	version := parts[len(parts)-2]
	product := strings.Join(parts[0:len(parts)-2], "-")
	fmt.Printf("product:%+v version:%+v spec:%+v \n", product, version, spec)
	uploadSpec(product, version, style, path)
	return nil
}

func uploadSpec(product, version, style, path string) error {
	ctx := context.TODO()
	// does the API exist? if not, create it
	{
		request := &rpcpb.GetProductRequest{}
		request.Name = "projects/atlas/products/" + product
		_, err := flameClient.GetProduct(ctx, request)
		if notFound(err) {
			request := &rpcpb.CreateProductRequest{}
			request.Parent = "projects/atlas"
			request.ProductId = product
			request.Product = &rpcpb.Product{}
			response, err := flameClient.CreateProduct(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else {
				log.Printf("failed to create %s/products/%s: %s",
					request.Parent, request.ProductId, err.Error())
			}
		}
	}
	// does the version exist? if not create it
	{
		request := &rpcpb.GetVersionRequest{}
		request.Name = "projects/atlas/products/" + product + "/versions/" + version
		_, err := flameClient.GetVersion(ctx, request)
		if notFound(err) {
			request := &rpcpb.CreateVersionRequest{}
			request.Parent = "projects/atlas/products/" + product
			request.VersionId = version
			request.Version = &rpcpb.Version{}
			response, err := flameClient.CreateVersion(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else {
				log.Printf("failed to create %s/versions/%s: %s",
					request.Parent, request.VersionId, err.Error())
			}
		}
	}
	// does the spec exist? if not, create it
	{
		filename := filepath.Base(path)

		request := &rpcpb.GetSpecRequest{}
		request.Name = "projects/atlas/products/" + product +
			"/versions/" + version +
			"/specs/" + filename
		_, err := flameClient.GetSpec(ctx, request)
		if notFound(err) {
			fileBytes, err := ioutil.ReadFile(path)
			// gzip the bytes
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
			request.Parent = "projects/atlas/products/" + product +
				"/versions/" + version
			request.SpecId = filename
			request.Spec = &rpcpb.Spec{}
			request.Spec.Style = style
			request.Spec.Filename = filename
			request.Spec.Contents = buf.Bytes()
			response, err := flameClient.CreateSpec(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else {
				details := fmt.Sprintf("contents-length: %d", len(request.Spec.Contents))
				log.Printf("failed to create %s/specs/%s: %s [%s]",
					request.Parent, request.SpecId, err.Error(), details)
			}
		}
	}
	return nil
}
