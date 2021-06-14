// Copyright 2021 Google LLC. All Rights Reserved.
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

package upload

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	uploadCsvCmd.Flags().String("project_id", "", "Project id.")
	uploadCsvCmd.MarkFlagRequired("project_id")
	uploadCsvCmd.Flags().String("delimiter", ",", "Field delimiter of the CSV file.")
}

var uploadCsvCmd = &cobra.Command{
	Use:   "csv file --project_id=value [--delimiter=value]",
	Short: "Upload API specs from a CSV file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		flagset := cmd.LocalFlags()
		projectID, err := flagset.GetString("project_id")
		if err != nil {
			log.Fatalf("Failed to get project_id string from flags: %s", err)
		}

		delimiter, err := flagset.GetString("delimiter")
		if err != nil {
			log.Fatalf("Failed to get delimiter string from flags: %s", err)
		} else if len(delimiter) != 1 {
			log.Fatalf("Invalid delimiter %q: must be exactly one character", delimiter)
		}

		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("Failed to create client: %s", err)
		}
		core.EnsureProjectExists(ctx, client, projectID)

		taskQueue := make(chan core.Task, 64)
		for i := 0; i < 64; i++ {
			core.WaitGroup().Add(1)
			go core.Worker(ctx, taskQueue)
		}
		defer core.WaitGroup().Wait()
		defer close(taskQueue)

		file, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("Failed to open file: %s", err)
		}
		defer file.Close()

		r := uploadCSVReader{
			Reader:    csv.NewReader(file),
			Delimiter: rune(delimiter[0]),
		}

		for row, err := r.Read(); err != io.EOF; row, err = r.Read() {
			if err != nil {
				log.Fatalf("Failed to read row from file: %s", err)
			}

			taskQueue <- &uploadSpecTask{
				ctx:       ctx,
				client:    client,
				projectID: projectID,
				apiID:     row.ApiID,
				versionID: row.VersionID,
				specID:    row.SpecID,
				filepath:  row.Filepath,
			}
		}
	},
}

type uploadCSVReader struct {
	Reader      *csv.Reader
	Delimiter   rune
	columnIndex map[string]int
}

type uploadCSVRow struct {
	ApiID     string
	VersionID string
	SpecID    string
	Filepath  string
}

func (r *uploadCSVReader) Read() (uploadCSVRow, error) {
	r.Reader.Comma = r.Delimiter
	row, err := r.Reader.Read()
	if err != nil {
		return uploadCSVRow{}, err
	}

	// Build an index of { column name -> column number } using the header row.
	if r.columnIndex == nil {
		if err := r.buildColumnIndex(row); err != nil {
			return uploadCSVRow{}, err
		}

		// Return the first non-header row instead of the header itself.
		return r.Read()
	}

	return uploadCSVRow{
		ApiID:     row[r.columnIndex["api_id"]],
		VersionID: row[r.columnIndex["version_id"]],
		SpecID:    row[r.columnIndex["spec_id"]],
		Filepath:  row[r.columnIndex["filepath"]],
	}, nil
}

func (r *uploadCSVReader) buildColumnIndex(header []string) error {
	want := map[string]bool{
		"api_id":     true,
		"version_id": true,
		"spec_id":    true,
		"filepath":   true,
	}

	var errs []error
	r.columnIndex = make(map[string]int, len(header))
	for i, name := range header {
		r.columnIndex[name] = i
		if !want[name] {
			errs = append(errs, fmt.Errorf("unexpected column name %q", name))
		}
	}

	for name := range want {
		if _, got := r.columnIndex[name]; !got {
			errs = append(errs, fmt.Errorf("expected column name %q", name))
		}
	}

	if len(errs) > 0 {
		msg := fmt.Sprintf("invalid header %v:", header)
		for _, err := range errs {
			msg += "\n\t" + err.Error()
		}
		return errors.New(msg)
	}

	return nil
}

type uploadSpecTask struct {
	ctx       context.Context
	client    connection.Client
	projectID string
	apiID     string
	versionID string
	specID    string
	filepath  string
}

func (t uploadSpecTask) Run() error {
	api, err := t.client.CreateApi(t.ctx, &rpc.CreateApiRequest{
		Parent: fmt.Sprintf("projects/%s", t.projectID),
		ApiId:  t.apiID,
		Api:    &rpc.Api{},
	})

	switch status.Code(err) {
	case codes.OK:
		log.Printf("Created API: %s", api.GetName())
	case codes.AlreadyExists:
		api = &rpc.Api{
			Name: fmt.Sprintf("projects/%s/apis/%s", t.projectID, t.apiID),
		}
	default:
		return fmt.Errorf("failed to ensure API exists: %s", err)
	}

	version, err := t.client.CreateApiVersion(t.ctx, &rpc.CreateApiVersionRequest{
		Parent:       api.GetName(),
		ApiVersionId: t.versionID,
		ApiVersion:   &rpc.ApiVersion{},
	})

	switch status.Code(err) {
	case codes.OK:
		log.Printf("Created API version: %s", version.GetName())
	case codes.AlreadyExists:
		version = &rpc.ApiVersion{
			Name: fmt.Sprintf("projects/%s/apis/%s/versions/%s", t.projectID, t.apiID, t.versionID),
		}
	default:
		return fmt.Errorf("failed to ensure API version exists: %s", err)
	}

	contents, err := ioutil.ReadFile(t.filepath)
	if err != nil {
		return err
	}

	compressed, err := core.GZippedBytes(contents)
	if err != nil {
		return err
	}

	spec, err := t.client.CreateApiSpec(t.ctx, &rpc.CreateApiSpecRequest{
		Parent:    version.GetName(),
		ApiSpecId: t.specID,
		ApiSpec: &rpc.ApiSpec{
			// TODO: How do we choose a mime type?
			MimeType: core.OpenAPIMimeType("+gzip", "3.0.0"),
			Contents: compressed,
		},
	})

	switch status.Code(err) {
	case codes.OK:
		log.Printf("Created API spec: %s", spec.GetName())
	case codes.AlreadyExists:
		// When the spec already exists we can silently continue.
	default:
		return fmt.Errorf("failed to upload API spec: %s", err)
	}

	return nil
}

func (t uploadSpecTask) String() string {
	return fmt.Sprintf("upload spec %s", t.filepath)
}
