// Copyright 2022 Google LLC.
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
	"log"
	"sort"
	"strings"
	"testing"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
)

func TestProtos(t *testing.T) {
	const (
		projectID   = "protos-test"
		projectName = "projects/" + projectID
		parent      = projectName + "/locations/global"
	)
	// Create a registry client.
	ctx := context.Background()
	registryClient, adminClient := grpctest.SetupRegistry(ctx, t, projectID, nil)

	cmd := Command()
	args := []string{"protos", "testdata/protos", "--parent", parent}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %+v returned error: %s", args, err)
	}
	tests := []struct {
		desc              string
		spec              string
		wantProtoCount    int
		wantMetadataCount int
	}{
		{
			desc:              "Apigee Registry",
			spec:              "apis/apigeeregistry/versions/v1/specs/google-cloud-apigeeregistry-v1",
			wantProtoCount:    11,
			wantMetadataCount: 2,
		},
		{
			desc:              "Example Library",
			spec:              "apis/library-example/versions/v1/specs/google-example-library-v1",
			wantProtoCount:    6,
			wantMetadataCount: 3,
		},
	}
	for _, test := range tests {
		// Get the uploaded spec
		result, err := registryClient.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
			Name: "projects/" + projectID + "/locations/global/" + test.spec,
		})
		if err != nil {
			t.Fatalf("unable to fetch spec %s", test.spec)
		}
		// Verify the content type.
		if result.ContentType != "application/x.protobuf+zip" {
			t.Errorf("Invalid mime type for %s: %s", test.spec, result.ContentType)
		}
		// Verify that the zip contains the expected number of files
		m, err := compress.UnzipArchiveToMap(result.Data)
		if err != nil {
			t.Fatalf("unable to unzip spec %s", test.spec)
		}
		proto_filenames := make([]string, 0)
		metadata_filenames := make([]string, 0)
		for filename := range m {
			if strings.HasSuffix(filename, ".proto") {
				proto_filenames = append(proto_filenames, filename)
			} else {
				metadata_filenames = append(metadata_filenames, filename)
			}
		}
		if len(proto_filenames) != test.wantProtoCount {
			t.Errorf("Archive contains incorrect number of proto files (%d, expected %d)",
				len(proto_filenames),
				test.wantProtoCount)
			sort.Strings(proto_filenames)
			for i, s := range proto_filenames {
				log.Printf("%d: %s", i, s)
			}
		}
		if len(metadata_filenames) != test.wantMetadataCount {
			t.Errorf("Archive contains incorrect number of metadata files (%d, expected %d)",
				len(metadata_filenames),
				test.wantMetadataCount)
			sort.Strings(metadata_filenames)
			for i, s := range metadata_filenames {
				log.Printf("%d: %s", i, s)
			}
		}
	}
	// Delete the test project.
	req := &rpc.DeleteProjectRequest{
		Name:  projectName,
		Force: true,
	}
	err := adminClient.DeleteProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to delete test project: %s", err)
	}
}

func TestProtosMissingParent(t *testing.T) {
	const (
		projectID   = "missing"
		projectName = "projects/" + projectID
		parent      = projectName + "/locations/global"
	)
	tests := []struct {
		desc string
		args []string
	}{
		{
			desc: "parent",
			args: []string{"protos", "nonexistent-specs-dir", "--parent", parent},
		},
		{
			desc: "project-id",
			args: []string{"protos", "nonexistent-specs-dir", "--project-id", projectID},
		},
		{
			desc: "unspecified",
			args: []string{"protos", "nonexistent-specs-dir"},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs(test.args)
			if cmd.Execute() == nil {
				t.Error("expected error, none reported")
			}
		})
	}
}
