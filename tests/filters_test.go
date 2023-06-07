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

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
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
	t.Helper()
	if unavailable(err) {
		t.Logf("Unable to connect to registry server. Is it running?")
		t.FailNow()
	}
	if err != nil {
		t.Errorf(message, err.Error())
	}
}

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestFilters(t *testing.T) {
	// Define test patterns and expected outputs.
	// These maps will be applied to resources as labels, mapping one-to-one with test resources.
	testLabels := []map[string]string{
		{"a": "0", "b": "0", "c": "0"},
		{"a": "1", "b": "0", "c": "1"},
		{"a": "0", "b": "1"},
		{"a": "1", "b": "1"},
	}
	// Define test filters with their expected return counts (and error occurrences).
	// See https://github.com/google/cel-spec/blob/master/doc/langdef.md#field-selection for CEL details.
	testFilters := []struct {
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
	// Create a registry client.
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "filters", nil)

	// Create some sample apis.
	apiParent := "projects/filters/locations/global"
	for i := 0; i < len(testLabels); i++ {
		req := &rpc.CreateApiRequest{
			Parent: apiParent,
			ApiId:  fmt.Sprintf("%d", i),
			Api: &rpc.Api{
				Labels: testLabels[i],
			},
		}
		_, err := registryClient.CreateApi(ctx, req)
		check(t, "error creating api %s", err)
	}
	// List all of the apis.
	{
		req := &rpc.ListApisRequest{
			Parent: apiParent,
		}
		it := registryClient.ListApis(ctx, req)
		count, err := countApis(it)
		check(t, "error listing apis %s", err)
		if count != len(testLabels) {
			t.Errorf("Incorrect count %d, expected %d", count, len(testLabels))
		}
	}
	// Test api filters.
	for i := 0; i < len(testFilters); i++ {
		testFilter := testFilters[i]
		req := &rpc.ListApisRequest{
			Parent: apiParent,
			Filter: testFilter.filter,
		}
		it := registryClient.ListApis(ctx, req)
		count, err := countApis(it)
		if err != nil && !testFilter.expectError {
			t.Errorf("Filter %s produced an unexpected error %s", testFilter.filter, err.Error())
		}
		if err == nil && testFilter.expectError {
			t.Errorf("Filter %s was expected to produce an error and did not", testFilter.filter)
		}
		if count != testFilter.expectedCount {
			t.Errorf("Incorrect count for filter %s: %d, expected %d", testFilter.filter, count, testFilter.expectedCount)
		}
	}
	// Create some sample versions.
	versionParent := "projects/filters/locations/global/apis/1"
	for i := 0; i < len(testLabels); i++ {
		req := &rpc.CreateApiVersionRequest{
			Parent:       versionParent,
			ApiVersionId: fmt.Sprintf("%d", i),
			ApiVersion: &rpc.ApiVersion{
				Labels: testLabels[i],
			},
		}
		_, err := registryClient.CreateApiVersion(ctx, req)
		check(t, "error creating version %s", err)
	}
	// List all of the versions.
	{
		req := &rpc.ListApiVersionsRequest{
			Parent: versionParent,
		}
		it := registryClient.ListApiVersions(ctx, req)
		count, err := countApiVersions(it)
		check(t, "error listing api versions %s", err)
		if count != len(testLabels) {
			t.Errorf("Incorrect count %d, expected %d", count, len(testLabels))
		}
	}
	// Test version filters.
	for i := 0; i < len(testFilters); i++ {
		testFilter := testFilters[i]
		req := &rpc.ListApiVersionsRequest{
			Parent: versionParent,
			Filter: testFilter.filter,
		}
		it := registryClient.ListApiVersions(ctx, req)
		count, err := countApiVersions(it)
		if err != nil && !testFilter.expectError {
			t.Errorf("Filter %s produced an unexpected error %s", testFilter.filter, err.Error())
		}
		if err == nil && testFilter.expectError {
			t.Errorf("Filter %s was expected to produce an error and did not", testFilter.filter)
		}
		if count != testFilter.expectedCount {
			t.Errorf("Incorrect count for filter %s: %d, expected %d", testFilter.filter, count, testFilter.expectedCount)
		}
	}
	// Create some sample specs.
	specParent := "projects/filters/locations/global/apis/1/versions/1"
	for i := 0; i < len(testLabels); i++ {
		req := &rpc.CreateApiSpecRequest{
			Parent:    specParent,
			ApiSpecId: fmt.Sprintf("000%d", i),
			ApiSpec: &rpc.ApiSpec{
				Labels: testLabels[i],
			},
		}
		_, err := registryClient.CreateApiSpec(ctx, req)
		check(t, "error creating spec %s", err)
	}
	// List all of the specs.
	{
		req := &rpc.ListApiSpecsRequest{
			Parent: specParent,
		}
		it := registryClient.ListApiSpecs(ctx, req)
		count, err := countApiSpecs(it)
		check(t, "error listing specs %s", err)
		if count != len(testLabels) {
			t.Errorf("Incorrect count %d, expected %d", count, len(testLabels))
		}
	}
	// Test spec filters.
	for i := 0; i < len(testFilters); i++ {
		testFilter := testFilters[i]
		req := &rpc.ListApiSpecsRequest{
			Parent: specParent,
			Filter: testFilter.filter,
		}
		it := registryClient.ListApiSpecs(ctx, req)
		count, err := countApiSpecs(it)
		if err != nil && !testFilter.expectError {
			t.Errorf("Filter %s produced an unexpected error %s", testFilter.filter, err.Error())
		}
		if err == nil && testFilter.expectError {
			t.Errorf("Filter %s was expected to produce an error and did not", testFilter.filter)
		}
		if count != testFilter.expectedCount {
			t.Errorf("Incorrect count for filter %s: %d, expected %d", testFilter.filter, count, testFilter.expectedCount)
		}
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

func countApiVersions(it *gapic.ApiVersionIterator) (int, error) {
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

func countApiSpecs(it *gapic.ApiSpecIterator) (int, error) {
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
