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
	"path/filepath"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
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
	fileBytes, _ := ioutil.ReadFile(filename)
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, err := zw.Write(fileBytes)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

func TestCRUD(t *testing.T) {
	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Clear the test project.
	{
		req := &rpc.DeleteProjectRequest{
			Name: "projects/test",
		}
		err = registryClient.DeleteProject(ctx, req)
		if status.Code(err) != codes.NotFound {
			check(t, "Failed to delete test project: %+v", err)
		}
	}
	// Create the test project.
	{
		req := &rpc.CreateProjectRequest{
			ProjectId: "test",
			Project: &rpc.Project{
				DisplayName: "Test",
				Description: "A test catalog",
			},
		}
		project, err := registryClient.CreateProject(ctx, req)
		check(t, "error creating project %s", err)
		if project.GetName() != "projects/test" {
			t.Errorf("Invalid project name %s", project.GetName())
		}
	}
	// Create a map to use for labels and annotations.
	sampleMap := map[string]string{"one": "1", "two": "2", "three": "3"}
	// Create the sample api.
	{
		req := &rpc.CreateApiRequest{
			Parent: "projects/test/locations/global",
			ApiId:  "sample",
			Api: &rpc.Api{
				DisplayName:  "Sample",
				Description:  "A sample API",
				Availability: "GENERAL",
				Labels:       sampleMap,
				Annotations:  sampleMap,
			},
		}
		_, err := registryClient.CreateApi(ctx, req)
		check(t, "error creating api %s", err)
	}
	// Create the sample 1.0.0 version.
	{
		req := &rpc.CreateApiVersionRequest{
			Parent:       "projects/test/locations/global/apis/sample",
			ApiVersionId: "1.0.0",
			ApiVersion: &rpc.ApiVersion{
				Labels:      sampleMap,
				Annotations: sampleMap,
			},
		}
		_, err := registryClient.CreateApiVersion(ctx, req)
		check(t, "error creating version %s", err)
	}
	// Upload the sample 1.0.0 OpenAPI spec.
	{
		buf, err := readAndGZipFile(filepath.Join("testdata", "openapi.yaml"))
		check(t, "error reading spec", err)
		req := &rpc.CreateApiSpecRequest{
			Parent:    "projects/test/locations/global/apis/sample/versions/1.0.0",
			ApiSpecId: "openapi.yaml",
			ApiSpec: &rpc.ApiSpec{
				MimeType:    "application/x.openapi+gzip;version=3.0.0",
				Contents:    buf.Bytes(),
				Labels:      sampleMap,
				Annotations: sampleMap,
			},
		}
		_, err = registryClient.CreateApiSpec(ctx, req)
		check(t, "error creating spec %s", err)
	}
	// Check the created API.
	{
		req := &rpc.GetApiRequest{
			Name: "projects/test/locations/global/apis/sample",
		}
		api, err := registryClient.GetApi(ctx, req)
		check(t, "error getting api %s", err)
		if diff := cmp.Diff(api.GetLabels(), sampleMap); diff != "" {
			t.Errorf("Unexpected api labels %+v", api.GetLabels())
		}
		if diff := cmp.Diff(api.GetAnnotations(), sampleMap); diff != "" {
			t.Errorf("Unexpected api annotations %+v", api.GetAnnotations())
		}
	}
	// Check the created version.
	{
		req := &rpc.GetApiVersionRequest{
			Name: "projects/test/locations/global/apis/sample/versions/1.0.0",
		}
		version, err := registryClient.GetApiVersion(ctx, req)
		check(t, "error getting version %s", err)
		if diff := cmp.Diff(version.GetLabels(), sampleMap); diff != "" {
			t.Errorf("Unexpected version labels %+v", version.GetLabels())
		}
		if diff := cmp.Diff(version.GetAnnotations(), sampleMap); diff != "" {
			t.Errorf("Unexpected version annotations %+v", version.GetAnnotations())
		}
	}
	// Check the created spec.
	var revision string
	{
		req := &rpc.GetApiSpecRequest{
			Name: "projects/test/locations/global/apis/sample/versions/1.0.0/specs/openapi.yaml",
		}
		spec, err := registryClient.GetApiSpec(ctx, req)
		check(t, "error getting spec %s", err)
		if diff := cmp.Diff(spec.GetLabels(), sampleMap); diff != "" {
			t.Errorf("Unexpected spec labels %+v", spec.GetLabels())
		}
		if diff := cmp.Diff(spec.GetAnnotations(), sampleMap); diff != "" {
			t.Errorf("Unexpected spec annotations %+v", spec.GetAnnotations())
		}
		revision = spec.GetRevisionId()
	}
	// Compute some common values for subsequent tests.
	buf, err := ioutil.ReadFile(filepath.Join("testdata", "openapi.yaml"))
	check(t, "error reading spec", err)
	expectedHash := hashForBytes(buf)
	expectedContentType := "application/x.openapi;version=3.0.0"
	// Check the contents of the created spec.
	{
		req := &rpc.GetApiSpecContentsRequest{
			Name: "projects/test/locations/global/apis/sample/versions/1.0.0/specs/openapi.yaml",
		}
		response, err := registryClient.GetApiSpecContents(ctx, req)
		check(t, "error getting spec contents %s", err)
		if err == nil {
			if response.GetContentType() != expectedContentType {
				t.Errorf("Unexpected content type %q", response.GetContentType())
			}
			contentHash := hashForBytes(response.Data)
			if contentHash != expectedHash {
				t.Errorf("Contents failed to match %s != %s", contentHash, expectedHash)
			}
		}
	}
	// Check the contents of the created revision.
	{
		req := &rpc.GetApiSpecContentsRequest{
			Name: "projects/test/locations/global/apis/sample/versions/1.0.0/specs/openapi.yaml@" + revision,
		}
		response, err := registryClient.GetApiSpecContents(ctx, req)
		check(t, "error getting spec contents %s", err)
		if err == nil {
			if response.GetContentType() != expectedContentType {
				t.Errorf("Unexpected content type %q", response.GetContentType())
			}
			contentHash := hashForBytes(response.Data)
			if contentHash != expectedHash {
				t.Errorf("Contents failed to match %s != %s", contentHash, expectedHash)
			}
		}
	}
	// Tag the revision.
	revisionTag := "prod"
	{
		req := &rpc.TagApiSpecRevisionRequest{
			Name: "projects/test/locations/global/apis/sample/versions/1.0.0/specs/openapi.yaml@" + revision,
			Tag:  revisionTag,
		}
		_, err := registryClient.TagApiSpecRevision(ctx, req)
		check(t, "error tagging spec %s", err)
	}
	// Check the contents of the tagged revision.
	{
		req := &rpc.GetApiSpecContentsRequest{
			Name: "projects/test/locations/global/apis/sample/versions/1.0.0/specs/openapi.yaml@" + revisionTag,
		}
		response, err := registryClient.GetApiSpecContents(ctx, req)
		check(t, "error getting spec contents %s", err)
		if err == nil {
			if response.GetContentType() != expectedContentType {
				t.Errorf("Unexpected content type %q", response.GetContentType())
			}
			contentHash := hashForBytes(response.Data)
			if contentHash != expectedHash {
				t.Errorf("Contents failed to match %s != %s", contentHash, expectedHash)
			}
		}
	}
	testArtifacts(ctx, registryClient, t, "projects/test/locations/global")
	testArtifacts(ctx, registryClient, t, "projects/test/locations/global/apis/sample")
	testArtifacts(ctx, registryClient, t, "projects/test/locations/global/apis/sample/versions/1.0.0")
	testArtifacts(ctx, registryClient, t, "projects/test/locations/global/apis/sample/versions/1.0.0/specs/openapi.yaml")
	// Delete the test project.
	{
		req := &rpc.DeleteProjectRequest{
			Name: "projects/test",
		}
		err = registryClient.DeleteProject(ctx, req)
		check(t, "Failed to delete test project: %+v", err)
	}
}

func hashForBytes(b []byte) string {
	h := sha256.New()
	_, _ = h.Write(b)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// testArtifacts verifies artifact operations on a specified entity.
func testArtifacts(ctx context.Context, registryClient connection.Client, t *testing.T, parent string) {
	messageContents := []byte("hello")
	messageHash := hashForBytes(messageContents)
	messageLength := int32(len(messageContents))
	messageMimeType := "text/plain"
	// Set the artifact.
	{
		req := &rpc.CreateArtifactRequest{
			Parent:     parent,
			ArtifactId: "sample",
			Artifact: &rpc.Artifact{
				MimeType: messageMimeType,
				Contents: messageContents,
			},
		}
		_, err := registryClient.CreateArtifact(ctx, req)
		check(t, "error creating artifact %s", err)
	}
	// Check the artifact.
	{
		req := &rpc.GetArtifactRequest{
			Name: fmt.Sprintf("%s/artifacts/sample", parent),
		}
		resp, err := registryClient.GetArtifact(ctx, req)
		check(t, "error getting artifact %s", err)
		if resp.GetMimeType() != messageMimeType {
			t.Errorf("Unexpected mime type %s (expected %s)", resp.GetMimeType(), messageMimeType)
		}
		if resp.GetSizeBytes() != messageLength {
			t.Errorf("Unexpected length %d (expected %d)", resp.GetSizeBytes(), messageLength)
		}
		if resp.GetHash() != messageHash {
			t.Errorf("Unexpected hash value %s (expected %s)", resp.GetHash(), messageHash)
		}
	}
	// Check the artifact contents.
	{
		req := &rpc.GetArtifactContentsRequest{
			Name: fmt.Sprintf("%s/artifacts/sample", parent),
		}
		resp, err := registryClient.GetArtifactContents(ctx, req)
		check(t, "error getting artifact contents %s", err)
		if err == nil {
			if resp.GetContentType() != messageMimeType {
				t.Errorf("Unexpected mime type %s (expected %s)", resp.GetContentType(), messageMimeType)
			}
			if !bytes.Equal(resp.GetData(), messageContents) {
				t.Errorf("Unexpected data %s (expected %s)", string(resp.GetData()), string(messageContents))
			}
		}
	}
}
