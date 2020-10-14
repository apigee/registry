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
	"io/ioutil"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
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
		check(t, "Failed to delete test project: %+v", err)
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
	// Create the sample api.
	{
		req := &rpc.CreateApiRequest{
			Parent: "projects/test",
			ApiId:  "sample",
			Api: &rpc.Api{
				DisplayName:  "Sample",
				Description:  "A sample API",
				Availability: "GENERAL",
				Owner:        "Acme APIs",
			},
		}
		_, err := registryClient.CreateApi(ctx, req)
		check(t, "error creating api %s", err)
	}
	// Create the sample 1.0.0 version.
	{
		req := &rpc.CreateVersionRequest{
			Parent:    "projects/test/apis/sample",
			VersionId: "1.0.0",
		}
		_, err := registryClient.CreateVersion(ctx, req)
		check(t, "error creating version %s", err)
	}
	// Upload the sample 1.0.0 OpenAPI spec.
	{
		buf, err := readAndGZipFile("openapi.yaml@r0")
		check(t, "error reading spec", err)
		req := &rpc.CreateSpecRequest{
			Parent: "projects/test/apis/sample/versions/1.0.0",
			SpecId: "openapi.yaml",
			Spec: &rpc.Spec{
				Style:    "openapi/v3+gzip",
				Contents: buf.Bytes(),
			},
		}
		_, err = registryClient.CreateSpec(ctx, req)
		check(t, "error creating spec %s", err)
	}
	testProperties(ctx, registryClient, t, "projects/test")
	testProperties(ctx, registryClient, t, "projects/test/apis/sample")
	testProperties(ctx, registryClient, t, "projects/test/apis/sample/versions/1.0.0")
	testProperties(ctx, registryClient, t, "projects/test/apis/sample/versions/1.0.0/specs/openapi.yaml")
}

// testProperties verifies property operations on a specified entity.
func testProperties(ctx context.Context, registryClient connection.Client, t *testing.T, parent string) {
	// Set a string property.
	{
		req := &rpc.CreatePropertyRequest{
			Parent:     parent,
			PropertyId: "string",
			Property: &rpc.Property{
				Value: &rpc.Property_StringValue{StringValue: "testing"},
			},
		}
		_, err := registryClient.CreateProperty(ctx, req)
		check(t, "error creating property %s", err)
	}
	// Set an int64 property.
	{
		req := &rpc.CreatePropertyRequest{
			Parent:     parent,
			PropertyId: "int64",
			Property: &rpc.Property{
				Value: &rpc.Property_Int64Value{Int64Value: 123},
			},
		}
		_, err := registryClient.CreateProperty(ctx, req)
		check(t, "error creating property %s", err)
	}
	// Set a double property.
	{
		req := &rpc.CreatePropertyRequest{
			Parent:     parent,
			PropertyId: "double",
			Property: &rpc.Property{
				Value: &rpc.Property_DoubleValue{DoubleValue: 123.456},
			},
		}
		_, err := registryClient.CreateProperty(ctx, req)
		check(t, "error creating property %s", err)
	}
	// Set a boolean property.
	{
		req := &rpc.CreatePropertyRequest{
			Parent:     parent,
			PropertyId: "bool",
			Property: &rpc.Property{
				Value: &rpc.Property_BoolValue{BoolValue: true},
			},
		}
		_, err := registryClient.CreateProperty(ctx, req)
		check(t, "error creating property %s", err)
	}
	// Set a bytes property.
	if true {
		req := &rpc.CreatePropertyRequest{
			Parent:     parent,
			PropertyId: "bytes",
			Property: &rpc.Property{
				Value: &rpc.Property_BytesValue{BytesValue: []byte("hello")},
			},
		}
		_, err := registryClient.CreateProperty(ctx, req)
		check(t, "error creating property %s", err)
	}
	// Set a message property.
	if true {
		req := &rpc.CreatePropertyRequest{
			Parent:     parent,
			PropertyId: "message",
			Property: &rpc.Property{
				Value: &rpc.Property_MessageValue{MessageValue: &anypb.Any{TypeUrl: "echo", Value: []byte{
					0x0a, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
				}}},
			},
		}
		_, err := registryClient.CreateProperty(ctx, req)
		check(t, "error creating property %s", err)
	}
}
