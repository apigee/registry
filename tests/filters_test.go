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

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
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

func TestFilters(t *testing.T) {
	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Clear the filters project.
	{
		req := &rpc.DeleteProjectRequest{
			Name: "projects/filters",
		}
		err = registryClient.DeleteProject(ctx, req)
		check(t, "Failed to delete filters project: %+v", err)
	}
	// Create the filters project.
	{
		req := &rpc.CreateProjectRequest{
			ProjectId: "filters",
			Project: &rpc.Project{
				DisplayName: "Test Filters",
				Description: "A project for testing filtering",
			},
		}
		project, err := registryClient.CreateProject(ctx, req)
		check(t, "error creating project %s", err)
		if project.GetName() != "projects/filters" {
			t.Errorf("Invalid project name %s", project.GetName())
		}
	}
	// Declare maps to use as labels.
	labelMaps := []map[string]string{
		{"a": "0", "b": "0", "c": "0"},
		{"a": "1", "b": "0", "c": "1"},
		{"a": "0", "b": "1"},
		{"a": "1", "b": "1"},
	}
	// Create some sample apis.
	for i := 0; i < len(labelMaps); i++ {
		req := &rpc.CreateApiRequest{
			Parent: "projects/filters",
			ApiId:  fmt.Sprintf("%d", i),
			Api: &rpc.Api{
				Labels: labelMaps[i],
			},
		}
		_, err := registryClient.CreateApi(ctx, req)
		check(t, "error creating api %s", err)
	}
	// List all of the APIs.
	{
		req := &rpc.ListApisRequest{
			Parent: "projects/filters",
		}
		it := registryClient.ListApis(ctx, req)
		count, err := countApis(it)
		check(t, "error listing apis %s", err)
		if count != len(labelMaps) {
			t.Errorf("Incorrect count %d, expected %d", count, len(labelMaps))
		}
	}
	// test some filters against their expected return counts
	// see https://github.com/google/cel-spec/blob/master/doc/langdef.md#field-selection for CEL details.
	pairs := []struct {
		filter        string // the filter pattern
		expectedCount int    // number of expected responses
		expectError   bool   // true if we expect an error
	}{
		{"has(labels.a)", 4, false},
		{"has(labels.b)", 4, false},
		{"has(labels.c)", 2, false},
		{"labels.a == '1'", 2, false},
		{"labels.b == '1'", 2, false},
		// if a field isn't present, the filter fails with an error about the missing field
		{"labels.c == '1'", 0, true},
		{"has(labels.c) && labels.c == '1'", 1, false},
		{"labels.a == '1' && labels.b == '1'", 1, false},
	}
	for i := 0; i < len(pairs); i++ {
		req := &rpc.ListApisRequest{
			Parent: "projects/filters",
			Filter: pairs[i].filter,
		}
		it := registryClient.ListApis(ctx, req)
		count, err := countApis(it)
		if err != nil && !pairs[i].expectError {
			t.Errorf("Filter %s produced an unexpected error %s", pairs[i].filter, err.Error())
		}
		if err == nil && pairs[i].expectError {
			t.Errorf("Filter %s was expected to produce an error and did not", pairs[i].filter)
		}
		if count != pairs[i].expectedCount {
			t.Errorf("Incorrect count for filter %s: %d, expected %d", pairs[i].filter, count, pairs[i].expectedCount)
		}
	}
	// Delete the test project.
	if false {
		req := &rpc.DeleteProjectRequest{
			Name: "projects/filters",
		}
		err = registryClient.DeleteProject(ctx, req)
		check(t, "Failed to delete filters project: %+v", err)
	}
}

func countApis(it *gapic.ApiIterator) (int, error) {
	count := 0
	for {
		_, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}
