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
					handleSpec(path, "oas2")
					completions <- 1
				}()
			}
			if strings.HasSuffix(path, "openapi.yaml") {
				processes++
				go func() {
					handleSpec(path, "oas3")
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

func handleSpec(path string, format string) error {
	name := strings.TrimPrefix(path, directory)
	parts := strings.Split(name, "/")
	spec := parts[len(parts)-1]
	version := parts[len(parts)-2]
	product := strings.Join(parts[0:len(parts)-2], "-")
	fmt.Printf("product:%+v version:%+v spec:%+v \n", product, version, spec)
	uploadSpec(product, version, format, path)
	return nil
}

func uploadSpec(product, version, spec, path string) error {
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
		request := &rpcpb.GetSpecRequest{}
		request.Name = "projects/atlas/products/" + product +
			"/versions/" + version +
			"/specs/" + spec
		_, err := flameClient.GetSpec(ctx, request)
		if notFound(err) {
			request := &rpcpb.CreateSpecRequest{}
			request.Parent = "projects/atlas/products/" + product +
				"/versions/" + version
			request.SpecId = spec
			request.Spec = &rpcpb.Spec{}
			request.Spec.Style = spec
			response, err := flameClient.CreateSpec(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else {
				log.Printf("failed to create %s/specs/%s: %s",
					request.Parent, request.SpecId, err.Error())
			}
		}
	}
	// does the file exist? if not, create it
	{
		filename := filepath.Base(path)
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

		request := &rpcpb.GetFileRequest{}
		request.Name = "projects/atlas/products/" + product +
			"/versions/" + version +
			"/specs/" + spec +
			"/files/" + filename
		_, err = flameClient.GetFile(ctx, request)
		if notFound(err) {
			request := &rpcpb.CreateFileRequest{}
			request.Parent = "projects/atlas/products/" + product +
				"/versions/" + version + "/specs/" + spec
			request.FileId = filename
			request.File = &rpcpb.File{}
			request.File.Contents = buf.Bytes()
			response, err := flameClient.CreateFile(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else {
				details := fmt.Sprintf("contents-length: %d", len(request.File.Contents))
				log.Printf("failed to create %s/files/%s: %s [%s]",
					request.Parent, request.FileId, err.Error(), details)
			}
		}
	}
	return nil
}
