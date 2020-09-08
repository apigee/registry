// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

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

	"github.com/apigee/registry/connection"
	rpcpb "github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

func init() {
	uploadCmd.AddCommand(uploadOpenAPICmd)
	uploadOpenAPICmd.Flags().String("project_id", "", "Project id.")
}

// uploadOpenAPICmd represents the upload protos command
var uploadOpenAPICmd = &cobra.Command{
	Use:   "openapi",
	Short: "Upload OpenAPI descriptions of APIs.",
	Long:  "Upload OpenAPI descriptions of APIs.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flagset := cmd.LocalFlags()
		projectID, err := flagset.GetString("project_id")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		fmt.Printf("openapi called with args %+v and project_id %s\n", args, projectID)

		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		if client == nil {
			log.Fatalf("that's bad")
		}
		ensureProjectExists(ctx, client, projectID)

		for _, arg := range args {
			log.Printf("%+v", arg)
			scanDirectoryForOpenAPI(projectID, arg)
		}
	},
}

func scanDirectoryForOpenAPI(projectID, directory string) {
	ctx := context.TODO()

	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(-1)
	}

	jobQueue := make(chan Runnable, 1024)

	workerCount := 32
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, jobQueue)
	}

	// walk a directory hierarchy, uploading every API spec that matches a set of expected file names.
	err = filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if strings.HasSuffix(path, "swagger.yaml") || strings.HasSuffix(path, "swagger.json") {
				jobQueue <- &uploadOpenAPIRunnable{
					registryClient: registryClient,
					projectID:      projectID,
					path:           path,
					directory:      directory,
					style:          "openapi/v2",
				}
			} else if strings.HasSuffix(path, "openapi.yaml") || strings.HasSuffix(path, "openapi.json") {
				jobQueue <- &uploadOpenAPIRunnable{
					registryClient: registryClient,
					projectID:      projectID,
					path:           path,
					directory:      directory,
					style:          "openapi/v3",
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	close(jobQueue)
	wg.Wait()
}

type uploadOpenAPIRunnable struct {
	registryClient connection.Client
	projectID      string
	path           string
	directory      string
	style          string
}

func sanitize(name string) string {
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	return name
}

func (job *uploadOpenAPIRunnable) run() error {
	// Compute the API name from the path to the spec file.
	name := strings.TrimPrefix(job.path, job.directory+"/")
	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return fmt.Errorf("Invalid API path: %s", name)
	}
	spec := sanitize(parts[len(parts)-1])
	version := sanitize(parts[len(parts)-2])
	api := sanitize(strings.Join(parts[0:len(parts)-2], "-"))
	log.Printf("apis/%s/versions/%s/specs/%s\n", api, version, spec)
	// Upload the spec for the specified api and version
	style := job.style
	path := job.path
	ctx := context.TODO()
	projectID := job.projectID
	registryClient := job.registryClient
	api = strings.Replace(api, "/", "-", -1)
	// If the API does not exist, create it.
	{
		request := &rpcpb.GetApiRequest{}
		request.Name = "projects/" + projectID + "/apis/" + api
		_, err := registryClient.GetApi(ctx, request)
		if notFound(err) {
			request := &rpcpb.CreateApiRequest{}
			request.Parent = "projects/" + projectID
			request.ApiId = api
			request.Api = &rpcpb.Api{}
			request.Api.DisplayName = api
			response, err := registryClient.CreateApi(ctx, request)
			if err == nil {
				log.Printf("created %s", response.Name)
			} else if alreadyExists(err) {
				log.Printf("already exists %s/apis/%s", request.Parent, request.ApiId)
			} else {
				log.Printf("failed to create %s/apis/%s: %s",
					request.Parent, request.ApiId, err.Error())
			}
		} else if err != nil {
			return err
		}
	}
	// If the API version does not exist, create it.
	{
		request := &rpcpb.GetVersionRequest{}
		request.Name = "projects/" + projectID + "/apis/" + api + "/versions/" + version
		_, err := registryClient.GetVersion(ctx, request)
		if notFound(err) {
			request := &rpcpb.CreateVersionRequest{}
			request.Parent = "projects/" + projectID + "/apis/" + api
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
		request.Name = "projects/" + projectID + "/apis/" + api +
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
			request.Parent = "projects/" + projectID + "/apis/" + api +
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
