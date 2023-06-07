// Copyright 2021 Google LLC.
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
	"os"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	oas2 "github.com/google/gnostic/openapiv2"
	oas3 "github.com/google/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func csvCommand() *cobra.Command {
	var (
		delimiter string
		jobs      int
	)

	cmd := &cobra.Command{
		Use:   "csv FILE",
		Short: "Upload API descriptions from a CSV file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(delimiter) != 1 {
				return fmt.Errorf("invalid delimiter %q: must be exactly one character", delimiter)
			}
			ctx := cmd.Context()
			parent, err := getParent(cmd)
			if err != nil {
				return fmt.Errorf("failed to identify parent project (%s)", err)
			}
			parentName, err := names.ParseProjectWithLocation(parent)
			if err != nil {
				return fmt.Errorf("error parsing project name (%s)", err)
			}
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				return fmt.Errorf("error getting client (%s)", err)
			}
			if err := visitor.VerifyLocation(ctx, client, parent); err != nil {
				return fmt.Errorf("parent does not exist (%s)", err)
			}

			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()

			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file (%s)", err)
			}
			defer file.Close()

			r := uploadCSVReader{
				Reader:    csv.NewReader(file),
				Delimiter: rune(delimiter[0]),
			}

			for row, err := r.Read(); err != io.EOF; row, err = r.Read() {
				if err != nil {
					return fmt.Errorf("failed to read row from file (%s)", err)
				}

				taskQueue <- &uploadSpecTask{
					client:     client,
					parentName: parentName,
					apiID:      row.ApiID,
					versionID:  row.VersionID,
					specID:     row.SpecID,
					filepath:   row.Filepath,
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "project ID to use for each upload (deprecated)")
	cmd.Flags().StringVar(&parent, "parent", "", "parent for the upload (projects/PROJECT/locations/LOCATION)")
	cmd.Flags().StringVar(&delimiter, "delimiter", ",", "field delimiter for the CSV file")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
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
	client     connection.RegistryClient
	parentName names.Project
	apiID      string
	versionID  string
	specID     string
	filepath   string
}

func (t uploadSpecTask) Run(ctx context.Context) error {
	apiName := t.parentName.Api(t.apiID)
	api, err := t.client.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: apiName.Parent(),
		ApiId:  apiName.ApiID,
		Api:    &rpc.Api{},
	})

	switch status.Code(err) {
	case codes.OK:
		log.Debugf(ctx, "Created API: %s", api.GetName())
	case codes.AlreadyExists:
		api = &rpc.Api{
			Name: apiName.String(),
		}
	default:
		return fmt.Errorf("failed to ensure API exists: %s", err)
	}

	versionName := apiName.Version(t.versionID)
	version, err := t.client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		Parent:       versionName.Parent(),
		ApiVersionId: versionName.VersionID,
		ApiVersion:   &rpc.ApiVersion{},
	})

	switch status.Code(err) {
	case codes.OK:
		log.Debugf(ctx, "Created API version: %s", version.GetName())
	case codes.AlreadyExists:
		version = &rpc.ApiVersion{
			Name: versionName.String(),
		}
	default:
		return fmt.Errorf("failed to ensure API version exists: %s", err)
	}

	contents, err := os.ReadFile(t.filepath)
	if err != nil {
		return err
	}

	oasVer := "unknown"
	if doc, err := oas3.ParseDocument(contents); err == nil {
		oasVer = doc.Openapi
	} else if doc, err := oas2.ParseDocument(contents); err == nil {
		oasVer = doc.Swagger
	}
	oasMimeType := mime.OpenAPIMimeType("+gzip", oasVer)

	compressed, err := compress.GZippedBytes(contents)
	if err != nil {
		return err
	}

	specName := versionName.Spec(t.specID)
	spec, err := t.client.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
		Parent:    specName.Parent(),
		ApiSpecId: specName.SpecID,
		ApiSpec: &rpc.ApiSpec{
			MimeType: oasMimeType,
			Contents: compressed,
		},
	})

	switch status.Code(err) {
	case codes.OK:
		log.Debugf(ctx, "Created API spec: %s", spec.GetName())
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
