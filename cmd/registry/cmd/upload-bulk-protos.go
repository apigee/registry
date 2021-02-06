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
	uploadBulkProtosCmd.Flags().String("base_uri", "", "Base to use for setting source_uri fields of uploaded specs.")
}

var uploadBulkProtosCmd = &cobra.Command{
	Use:   "protos",
	Short: "Bulk-upload Protocol Buffer descriptions from a directory of specs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flagset := cmd.LocalFlags()
		projectID, err := flagset.GetString("project_id")
		if err != nil {
			log.Fatal(err.Error())
		}
		if projectID == "" {
			log.Fatalf("Please specify a project_id")
		}
		baseURI, err := flagset.GetString("base_uri")
		if err != nil {
			log.Fatal(err.Error())
		}
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}
		core.EnsureProjectExists(ctx, client, projectID)
		for _, arg := range args {
			scanDirectoryForProtos(ctx, client, projectID, baseURI, arg)
		}
	},
}

func scanDirectoryForProtos(ctx context.Context, client connection.Client, projectID, baseURI, directory string) {
	// create a queue for upload tasks and wait for the workers to finish after filling it.
	taskQueue := make(chan core.Task, 1024)
	for i := 0; i < 64; i++ {
		core.WaitGroup().Add(1)
		go core.Worker(ctx, taskQueue)
	}
	defer core.WaitGroup().Wait()
	defer close(taskQueue)

	dirPattern := regexp.MustCompile("v.*[1-9]+.*")
	if err := filepath.Walk(directory, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-matching directories.
		filename := path.Base(filepath)
		if !info.IsDir() || !dirPattern.MatchString(filename) {
			return nil
		}

		taskQueue <- &uploadProtoTask{
			ctx:       ctx,
			client:    client,
			baseURI:   baseURI,
			projectID: projectID,
			path:      filepath,
			directory: directory,
		}

		return nil
	}); err != nil {
		log.Println(err)
	}
}

type uploadProtoTask struct {
	ctx       context.Context
	client    connection.Client
	baseURI   string
	projectID string
	path      string
	directory string
	apiID     string // computed at runtime
	apiOwner  string // computed at runtime
	versionID string // computed at runtime
	specID    string // computed at runtime
}

func (task *uploadProtoTask) Name() string {
	return "upload proto " + task.path
}

func (task *uploadProtoTask) Run() error {
	// Populate API path fields using the file's path.
	task.populateFields()
	log.Printf("^^ apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)

	// If the API does not exist, create it.
	if err := task.createAPI(); err != nil {
		return err
	}
	// If the API version does not exist, create it.
	if err := task.createVersion(); err != nil {
		return err
	}
	// If the API spec does not exist, create it.
	if err := task.createSpec(); err != nil {
		return err
	}
	return nil
}

func (task *uploadProtoTask) populateFields() {
	parts := strings.Split(task.apiPath(), "/")
	apiParts := parts[0 : len(parts)-1]

	task.apiID = strings.ReplaceAll(strings.Join(apiParts, "-"), "/", "-")
	task.apiOwner = strings.ReplaceAll(task.apiID, "-", "/")
	task.versionID = parts[len(parts)-1]
	task.specID = task.fileName()
}

func (task *uploadProtoTask) createAPI() error {
	if _, err := task.client.GetApi(task.ctx, &rpcpb.GetApiRequest{
		Name: task.apiName(),
	}); !core.NotFound(err) {
		return err // Returns nil when API is found without error.
	}

	response, err := task.client.CreateApi(task.ctx, &rpcpb.CreateApiRequest{
		Parent: task.projectName(),
		ApiId:  task.apiID,
		Api: &rpcpb.Api{
			Owner: task.apiOwner,
		},
	})
	if err != nil {
		log.Printf("error %s: %s", task.apiName(), err.Error())
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) createVersion() error {
	if _, err := task.client.GetVersion(task.ctx, &rpcpb.GetVersionRequest{
		Name: task.versionName(),
	}); !core.NotFound(err) {
		return err // Returns nil when version is found without error.
	}

	response, err := task.client.CreateVersion(task.ctx, &rpcpb.CreateVersionRequest{
		Parent:    task.apiName(),
		VersionId: task.versionID,
		Version:   &rpcpb.Version{},
	})
	if err != nil {
		log.Printf("error %s: %s", task.versionName(), err.Error())
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) createSpec() error {
	contents, err := task.zipContents()
	if err != nil {
		return err
	}

	if _, err := task.client.GetSpec(task.ctx, &rpcpb.GetSpecRequest{
		Name: task.specName(),
	}); !core.NotFound(err) {
		return err // Returns nil when spec is found without error.
	}

	request := &rpcpb.CreateSpecRequest{
		Parent: task.versionName(),
		SpecId: task.fileName(),
		Spec: &rpcpb.Spec{
			Style:    "proto+zip",
			Filename: task.fileName(),
			Contents: contents,
		},
	}
	if task.baseURI != "" {
		request.Spec.SourceUri = fmt.Sprintf("%s/%s", task.baseURI, task.apiPath())
	}

	response, err := task.client.CreateSpec(task.ctx, request)
	if err != nil {
		log.Printf("error %s: %s [contents-length: %d]", task.specName(), err.Error(), len(contents))
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) projectName() string {
	return fmt.Sprintf("projects/%s", task.projectID)
}

func (task *uploadProtoTask) apiName() string {
	return fmt.Sprintf("%s/apis/%s", task.projectName(), task.apiID)
}

func (task *uploadProtoTask) versionName() string {
	return fmt.Sprintf("%s/versions/%s", task.apiName(), task.versionID)
}

func (task *uploadProtoTask) specName() string {
	return fmt.Sprintf("%s/specs/%s", task.versionName(), task.specID)
}

func (task *uploadProtoTask) apiPath() string {
	prefix := task.directory + "/"
	return strings.TrimPrefix(task.path, prefix)
}

func (task *uploadProtoTask) fileName() string {
	return "protos.zip"
}

func (task *uploadProtoTask) zipContents() ([]byte, error) {
	prefix := task.directory + "/"
	contents, err := core.ZipArchiveOfPath(task.path, prefix)
	if err != nil {
		return nil, err
	}

	return contents.Bytes(), nil
}
