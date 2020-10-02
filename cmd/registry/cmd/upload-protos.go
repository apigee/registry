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
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	rpcpb "github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

func init() {
	uploadBulkCmd.AddCommand(uploadBulkProtosCmd)
	uploadBulkProtosCmd.Flags().String("project_id", "", "Project id.")
}

var uploadBulkProtosCmd = &cobra.Command{
	Use:   "protos",
	Short: "Bulk-upload Protocol Buffer descriptions of APIs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flagset := cmd.LocalFlags()
		projectID, err := flagset.GetString("project_id")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		if projectID == "" {
			log.Fatalf("Please specify a project_id")
		}
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		core.EnsureProjectExists(ctx, client, projectID)
		for _, arg := range args {
			scanDirectoryForProtos(ctx, client, projectID, arg)
		}
	},
}

func scanDirectoryForProtos(ctx context.Context, client connection.Client, projectID, directory string) {
	var err error

	r := regexp.MustCompile("v.*[1-9]+.*")

	taskQueue := make(chan core.Task, 1024)

	workerCount := 64
	for i := 0; i < workerCount; i++ {
		core.WaitGroup().Add(1)
		go core.Worker(ctx, taskQueue)
	}

	// walk a directory hierarchy, uploading every API spec that matches a set of expected file names.
	err = filepath.Walk(directory,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil // skip files
			}
			b := path.Base(p)
			if !r.MatchString(b) {
				return nil
			}
			// we need to upload this API spec
			taskQueue <- &uploadProtoTask{
				ctx:       ctx,
				client:    client,
				projectID: projectID,
				path:      p,
				directory: directory,
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	close(taskQueue)
	core.WaitGroup().Wait()
}

type uploadProtoTask struct {
	ctx       context.Context
	client    connection.Client
	projectID string
	path      string
	directory string
	apiID     string // computed at runtime
	versionID string // computed at runtime
	specID    string // computed at runtime
}

func (task *uploadProtoTask) Run() error {
	var err error
	// Compute the API name from the path to the spec file.
	prefix := task.directory + "/"
	name := strings.TrimPrefix(task.path, prefix)
	parts := strings.Split(name, "/")
	task.apiID = strings.Join(parts[0:len(parts)-1], "-")
	task.apiID = strings.Replace(task.apiID, "/", "-", -1)
	task.versionID = parts[len(parts)-1]
	log.Printf("^^ apis/%s/versions/%s/specs/protos.zip\n", task.apiID, task.versionID)
	// If the API does not exist, create it.
	err = task.createAPI()
	if err != nil {
		return err
	}
	// If the API version does not exist, create it.
	err = task.createVersion()
	if err != nil {
		return err
	}
	// If the API spec does not exist, create it.
	return task.createSpec()
}

func (task *uploadProtoTask) createAPI() error {
	request := &rpcpb.GetApiRequest{
		Name: "projects/" + task.projectID + "/apis/" + task.apiID,
	}
	_, err := task.client.GetApi(task.ctx, request)
	if core.NotFound(err) {
		request := &rpcpb.CreateApiRequest{
			Parent: "projects/" + task.projectID,
			ApiId:  task.apiID,
			Api: &rpcpb.Api{
				DisplayName: task.apiID,
			},
		}
		response, err := task.client.CreateApi(task.ctx, request)
		if err == nil {
			log.Printf("created %s", response.Name)
		} else if core.AlreadyExists(err) {
			log.Printf("found %s/apis/%s", request.Parent, request.ApiId)
		} else {
			log.Printf("error %s/apis/%s: %s",
				request.Parent, request.ApiId, err.Error())
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (task *uploadProtoTask) createVersion() error {
	request := &rpcpb.GetVersionRequest{
		Name: "projects/" + task.projectID + "/apis/" + task.apiID + "/versions/" + task.versionID,
	}
	_, err := task.client.GetVersion(task.ctx, request)
	if core.NotFound(err) {
		request := &rpcpb.CreateVersionRequest{
			Parent:    "projects/" + task.projectID + "/apis/" + task.apiID,
			VersionId: task.versionID,
			Version:   &rpcpb.Version{},
		}
		response, err := task.client.CreateVersion(task.ctx, request)
		if err == nil {
			log.Printf("created %s", response.Name)
		} else if core.AlreadyExists(err) {
			log.Printf("found %s/versions/%s", request.Parent, request.VersionId)
		} else {
			log.Printf("error %s/versions/%s: %s",
				request.Parent, request.VersionId, err.Error())
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (task *uploadProtoTask) createSpec() error {
	filename := "protos.zip"
	request := &rpcpb.GetSpecRequest{
		Name: "projects/" + task.projectID + "/apis/" + task.apiID +
			"/versions/" + task.versionID +
			"/specs/" + filename,
	}
	_, err := task.client.GetSpec(task.ctx, request)
	if core.NotFound(err) {
		prefix := task.directory + "/"
		// build a zip archive with the contents of the path
		// https://golangcode.com/create-zip-files-in-go/
		buf, err := core.ZipArchiveOfPath(task.path, prefix)
		if err != nil {
			return err
		}
		request := &rpcpb.CreateSpecRequest{
			Parent: "projects/" + task.projectID + "/apis/" + task.apiID + "/versions/" + task.versionID,
			SpecId: filename,
			Spec: &rpcpb.Spec{
				Style:    "proto+zip",
				Filename: "protos.zip",
				Contents: buf.Bytes(),
			},
		}
		response, err := task.client.CreateSpec(task.ctx, request)
		if err == nil {
			log.Printf("created %s", response.Name)
		} else if core.AlreadyExists(err) {
			log.Printf("found %s/specs/%s", request.Parent, request.SpecId)
		} else {
			details := fmt.Sprintf("contents-length: %d", len(request.Spec.Contents))
			log.Printf("error %s/specs/%s: %s [%s]",
				request.Parent, request.SpecId, err.Error(), details)
		}

	} else if err != nil {
		return err
	}
	return nil
}
