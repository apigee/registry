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

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func unavailable(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.Unavailable
}

func check(t *testing.T, message string, err error) {
	if unavailable(err) {
		t.Logf("Unable to connect to registry server. Is it running?")
		t.FailNow()
	}
	if err != nil {
		t.Errorf(message, err.Error())
	}
}

func readAndGZipFile(filename string) (*bytes.Buffer, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, err = zw.Write(fileBytes)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

func hashForBytes(b []byte) string {
	h := sha256.New()
	h.Write(b)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func listAllSpecs(ctx context.Context, registryClient connection.Client) []*rpc.ApiSpec {
	specs := make([]*rpc.ApiSpec, 0)
	req := &rpc.ListApiSpecsRequest{
		Parent: "projects/demo/apis/-/versions/-",
	}
	it := registryClient.ListApiSpecs(ctx, req)
	for {
		spec, err := it.Next()
		if err == nil {
			specs = append(specs, spec)
		} else {
			break
		}
	}
	return specs
}

func listAllSpecRevisionIDs(ctx context.Context, registryClient connection.Client) []string {
	revisionIDs := make([]string, 0)
	req := &rpc.ListApiSpecRevisionsRequest{
		Name: "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml",
	}
	it := registryClient.ListApiSpecRevisions(ctx, req)
	for {
		spec, err := it.Next()
		if err == nil {
			revisionIDs = append(revisionIDs, spec.RevisionId)
		} else {
			break
		}
	}
	return revisionIDs
}

func TestDemo(t *testing.T) {
	var revisionIDs []string      // holds revision ids from queries
	var specs []*rpc.ApiSpec      // holds specs from queries
	var originalRevisionID string // revision id of first revision
	var originalHash string       // hash of first revision

	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Clear the demo project.
	{
		req := &rpc.DeleteProjectRequest{
			Name: "projects/demo",
		}
		err = registryClient.DeleteProject(ctx, req)
		if status.Code(err) != codes.NotFound {
			check(t, "Failed to delete demo project: %+v", err)
		}
	}
	// Create the demo project.
	{
		req := &rpc.CreateProjectRequest{
			ProjectId: "demo",
			Project: &rpc.Project{
				DisplayName: "Demo",
				Description: "A demo catalog",
			},
		}
		project, err := registryClient.CreateProject(ctx, req)
		check(t, "error creating project %s", err)
		if project.GetName() != "projects/demo" {
			t.Errorf("Invalid project name %s", project.GetName())
		}
	}
	// List the demo project. This should return exactly one result.
	{
		req := &rpc.ListProjectsRequest{
			Filter: "project_id == 'demo'",
		}
		count := 0
		it := registryClient.ListProjects(ctx, req)
		for {
			project, err := it.Next()
			if err == nil {
				if project.Name != "projects/demo" {
					t.Errorf("Invalid project name: %s", project.Name)
				}
				count++
			} else {
				break
			}
		}
		if count != 1 {
			t.Errorf("Invalid project count: %d", count)
		}
	}
	// Get the demo project.
	{
		req := &rpc.GetProjectRequest{
			Name: "projects/demo",
		}
		project, err := registryClient.GetProject(ctx, req)
		check(t, "error getting project %s", err)
		if project.Name != "projects/demo" {
			t.Errorf("Invalid project name: %s", project.Name)
		}
	}
	// Create the petstore api.
	{
		req := &rpc.CreateApiRequest{
			Parent: "projects/demo",
			ApiId:  "petstore",
			Api: &rpc.Api{
				DisplayName:  "Swagger Petstore",
				Description:  "A sample API",
				Availability: "GENERAL",
			},
		}
		_, err := registryClient.CreateApi(ctx, req)
		check(t, "error creating api %s", err)
	}
	// Create the petstore 1.0.0 version.
	{
		req := &rpc.CreateApiVersionRequest{
			Parent:       "projects/demo/apis/petstore",
			ApiVersionId: "1.0.0",
			ApiVersion:   &rpc.ApiVersion{},
		}
		_, err := registryClient.CreateApiVersion(ctx, req)
		check(t, "error creating version %s", err)
	}
	// Upload the petstore 1.0.0 OpenAPI spec.
	{
		buf, err := readAndGZipFile("testdata/openapi.yaml@r0")
		check(t, "error reading spec", err)
		req := &rpc.CreateApiSpecRequest{
			Parent:    "projects/demo/apis/petstore/versions/1.0.0",
			ApiSpecId: "openapi.yaml",
			ApiSpec: &rpc.ApiSpec{
				MimeType: "application/x.openapi+gzip;version=3.0.0",
				Contents: buf.Bytes(),
			},
		}
		got, err := registryClient.CreateApiSpec(ctx, req)
		check(t, "error creating spec %s", err)
		originalRevisionID = got.GetRevisionId()
		originalHash = got.GetHash()
	}
	// Update the OpenAPI spec three times with different revisions.
	for _, filename := range []string{
		"testdata/openapi.yaml@r1",
		"testdata/openapi.yaml@r2",
		"testdata/openapi.yaml@r3",
	} {
		buf, err := readAndGZipFile(filename)
		check(t, "error reading spec", err)
		req := &rpc.UpdateApiSpecRequest{
			ApiSpec: &rpc.ApiSpec{
				Name:     "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				Contents: buf.Bytes(),
			},
		}
		_, err = registryClient.UpdateApiSpec(ctx, req)
		check(t, "error updating spec %s", err)
	}
	// List the spec revisions.
	{
		revisionIDs = listAllSpecRevisionIDs(ctx, registryClient)
		if len(revisionIDs) != 4 {
			t.Errorf("Incorrect revision count: %d (if this is zero, be sure that all indexes are built)", len(revisionIDs))
		}
	}
	if len(revisionIDs) > 0 {
		req := &rpc.GetApiSpecRequest{
			Name: "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml" + "@" + revisionIDs[len(revisionIDs)-1],
		}
		spec, err := registryClient.GetApiSpec(ctx, req)
		check(t, "error getting spec %s", err)
		// compute the hash of the original file
		buf, err := readAndGZipFile("testdata/openapi.yaml@r0")
		check(t, "error reading spec", err)
		if hash := hashForBytes(buf.Bytes()); spec.GetHash() != hash {
			t.Errorf("Hash mismatch %s != %s", spec.GetHash(), hash)
		}
	}
	// List specs; there should be only one.
	{
		specs = listAllSpecs(ctx, registryClient)
		if len(specs) != 1 {
			t.Errorf("Incorrect spec count %d, expected exactly one:\n%+v", len(specs), specs)
		}
	}
	// tag a spec revision
	{
		req := &rpc.TagApiSpecRevisionRequest{
			Name: "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml" + "@" + originalRevisionID,
			Tag:  "og",
		}
		taggedSpec, err := registryClient.TagApiSpecRevision(ctx, req)
		check(t, "error tagging spec", err)
		if taggedSpec.Name != "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml@og" {
			t.Errorf("Incorrect name of tagged spec: %s", taggedSpec.Name)
		}
		if taggedSpec.Hash != originalHash {
			t.Errorf("Incorrect hash for tagged spec: %s", taggedSpec.Hash)
		}
	}
	// tag the tagged revision
	{
		req := &rpc.TagApiSpecRevisionRequest{
			Name: "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml@og",
			Tag:  "first",
		}
		taggedSpec, err := registryClient.TagApiSpecRevision(ctx, req)
		check(t, "error tagging spec", err)
		if taggedSpec.Name != "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml@first" {
			t.Errorf("Incorrect name of tagged spec: %s", taggedSpec.Name)
		}
		if taggedSpec.Hash != originalHash {
			t.Errorf("Incorrect hash for tagged spec: %s", taggedSpec.Hash)
		}
	}
	// get a spec by its tag
	{
		req := &rpc.GetApiSpecRequest{
			Name: "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml@first",
		}
		spec, err := registryClient.GetApiSpec(ctx, req)
		check(t, "error getting spec %s", err)
		if spec.Hash != originalHash {
			t.Errorf("Incorrect hash for spec retrieved by tag: %s", spec.Hash)
		}
	}
	// rollback a spec revision (this creates a new revision that's a copy)
	{
		req := &rpc.RollbackApiSpecRequest{
			Name:       "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			RevisionId: "og",
		}
		spec, err := registryClient.RollbackApiSpec(ctx, req)
		check(t, "error rolling back spec %s", err)
		if spec.Hash != originalHash {
			t.Errorf("Incorrect hash for rolled-back spec: %s", spec.Hash)
		}
	}
	// List specs; there should be only one.
	{
		specs = listAllSpecs(ctx, registryClient)
		if len(specs) != 1 {
			t.Errorf("Incorrect spec count: %d", len(specs))
		}
	}
	// list spec revisions, there should now be five
	{
		revisionIDs = listAllSpecRevisionIDs(ctx, registryClient)
		if len(revisionIDs) != 5 {
			t.Errorf("Incorrect revision count: %d", len(revisionIDs))
		}
	}
	// delete a spec revision
	{
		req := &rpc.DeleteApiSpecRevisionRequest{
			Name: "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml@og",
		}
		err := registryClient.DeleteApiSpecRevision(ctx, req)
		check(t, "error deleting spec revision %s", err)
	}
	// list specs, there should be only one
	{
		specs = listAllSpecs(ctx, registryClient)
		if len(specs) != 1 {
			t.Errorf("Incorrect spec count: %d", len(specs))
		}
	}
	// list spec revisions, there should be four
	{
		revisionIDs = listAllSpecRevisionIDs(ctx, registryClient)
		if len(revisionIDs) != 4 {
			t.Errorf("Incorrect revision count: %d", len(revisionIDs))
		}
	}
	// delete the spec
	{
		req := &rpc.DeleteApiSpecRequest{
			Name: "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		}
		err := registryClient.DeleteApiSpec(ctx, req)
		check(t, "error deleting spec %s", err)
	}
	// list spec revisions, there should be none
	{
		revisionIDs = listAllSpecRevisionIDs(ctx, registryClient)
		if len(revisionIDs) != 0 {
			t.Errorf("Incorrect revision count: %d", len(revisionIDs))
		}
	}
	// Delete the demo project.
	{
		req := &rpc.DeleteProjectRequest{
			Name: "projects/demo",
		}
		err = registryClient.DeleteProject(ctx, req)
		check(t, "Failed to delete demo project: %+v", err)
	}
}
