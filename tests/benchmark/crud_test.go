// Copyright 2022 Google LLC. All Rights Reserved.
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

package benchmark

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/genproto/protobuf/field_mask"
)

var (
	registryProject    string
	versionCount       = 1
	apiNamePrefix      = "test"
	apiNameStartOffset = 0
	specPath           string
	mimeType           string
	pauseDuration      int64
	pageSize           int
)

func init() {
	flag.StringVar(&registryProject, "registry_project", "bench", "Name of the Registry project")
	flag.IntVar(&versionCount, "version_count", 1, "Number of versions of the API to create (Each version will have 1 spec)")
	flag.StringVar(&apiNamePrefix, "api_prefix", "test", "Prefix to use for the APIs")
	flag.IntVar(&apiNameStartOffset, "api_offset", 0, "Offset the name of the API with this value")
	flag.StringVar(&specPath, "spec_path", filepath.Join("testdata", "openapi.yaml"), "Absolute path to the specification file to use")
	flag.StringVar(&mimeType, "mime_type", "application/x.openapi;version=3.0.0", "MIME type of the spec file")
	flag.Int64Var(&pauseDuration, "pause_duration", 5, "Number of seconds to pause after every 1000 runs")
	flag.IntVar(&pageSize, "page_size", 20, "Page size for the List apis")
}

func getRootResource(b *testing.B) string {
	b.Helper()
	return fmt.Sprintf("projects/%s/locations/global", registryProject)
}

func getApiName(i int, b *testing.B) string {
	b.Helper()
	return fmt.Sprintf("%s-%d", apiNamePrefix, i+apiNameStartOffset)
}

func pauseAfter1000Iterations(i int, b *testing.B) {
	b.Helper()
	if i%1000 == 0 {
		/*
			Pause 5 seconds every 1000 iterations to reduce chances of failure on
			CREATE, UPDATE, DELETE
		*/
		time.Sleep(time.Duration(pauseDuration * time.Second.Nanoseconds()))
	}
}

func BenchmarkCreateApi(b *testing.B) {

	rootResource := getRootResource(b)
	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		req := &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  apiId,
			Api: &rpc.Api{
				DisplayName: apiId,
				Description: fmt.Sprintf("Description for %s", apiId),
			},
		}
		b.StartTimer()
		_, err := client.CreateApi(ctx, req)
		b.StopTimer()
		/**
		Go Benchmarking always runs all the benchmark tests once before it runs
		for the duration specified. So the first row always returns error.
		https://github.com/golang/go/issues/32051
		Fixes --benchtime=1x but does not fix --benchtime=100X
		https://github.com/golang/go/commit/b69f823ece741f21d06591657f4e0a5b17d492e3
		*/
		if i > 1 && err != nil {
			b.Errorf("CreateApi(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkCreateApiVersion(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		for j := 1; j <= versionCount; j++ {
			req := &rpc.CreateApiVersionRequest{
				Parent: fmt.Sprintf("%s/apis/%s", rootResource, apiId),
				ApiVersion: &rpc.ApiVersion{
					DisplayName: fmt.Sprintf("v%d", j),
					Description: fmt.Sprintf("v%d of %s", j, apiId),
				},
				ApiVersionId: fmt.Sprintf("v%d", j),
			}
			// Only take reading for the first version
			if j == 1 {
				b.StartTimer()
			}
			_, err := client.CreateApiVersion(ctx, req)
			if j == 1 {
				b.StopTimer()
			}
			if i > 1 && err != nil {
				b.Errorf("CreateApiVersion(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkCreateApiSpec(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()
	contents, err := ioutil.ReadFile(specPath)
	if err != nil {
		b.Errorf("BenchmarkCreateApiSpec could not read %s : %s", specPath, err)
		return
	}
	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		for j := 1; j <= versionCount; j++ {

			req := &rpc.CreateApiSpecRequest{
				Parent: fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, j),
				ApiSpec: &rpc.ApiSpec{
					MimeType:    mimeType,
					Contents:    contents,
					Filename:    fmt.Sprintf("spec-%s-v%d-spec%s", apiId, j, filepath.Ext(specPath)),
					Description: fmt.Sprintf("Spec for api[%s] version[v%d]", apiId, j),
				},
				ApiSpecId: fmt.Sprintf("spec-%d", j),
			}
			// Only take reading for the first version
			if j == 1 {
				b.StartTimer()
			}
			_, err = client.CreateApiSpec(ctx, req)
			if j == 1 {
				b.StopTimer()
			}
			if i > 1 && err != nil {
				b.Errorf("CreateApiSpec(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkCreateArtifact(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		for j := 1; j <= versionCount; j++ {
			artifactParent := fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, j, j)

			messageContents := []byte(artifactParent)
			messageMimeType := "text/plain"
			// Set the artifact.
			req := &rpc.CreateArtifactRequest{
				Parent:     artifactParent,
				ArtifactId: "self-parent-link",
				Artifact: &rpc.Artifact{
					MimeType: messageMimeType,
					Contents: messageContents,
				},
			}
			// Only take reading for the first version
			if j == 1 {
				b.StartTimer()
			}
			_, err = client.CreateArtifact(ctx, req)
			// Only take reading for the first version
			if j == 1 {
				b.StopTimer()
			}
			if i > 1 && err != nil {
				b.Errorf("CreateArtifact(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkUpdateApi(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		req := &rpc.GetApiRequest{
			Name: fmt.Sprintf("%s/apis/%s", rootResource, apiId),
		}
		api, err := client.GetApi(ctx, req)
		if err != nil {
			b.Errorf("BenchmarkUpdateApi:GetApi(%+v) returned unexpected error: %s", req, err)
		} else {
			api.DisplayName = fmt.Sprintf("Updated %s", apiId)
			updateReq := &rpc.UpdateApiRequest{
				Api: api,
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{"display_name"},
				},
			}

			_, err = client.UpdateApi(ctx, updateReq)

			/**
			Go Benchmarking always runs all the benchmark tests once before it runs
			for the duration specified. So the first row always returns error.
			*/
			if i > 1 && err != nil {
				b.Errorf("UpdateApi(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkUpdateApiVersion(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		//Update the first api version of every api
		req := &rpc.UpdateApiVersionRequest{
			ApiVersion: &rpc.ApiVersion{
				Name:        fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, 1),
				DisplayName: fmt.Sprintf("Updated v%d", 1),
				Description: fmt.Sprintf("Updated description v%d of %s", 1, apiId),
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"display_name", "description"},
			},
		}
		b.StartTimer()
		_, err := client.UpdateApiVersion(ctx, req)

		b.StopTimer()
		if i > 1 && err != nil {
			b.Errorf("UpdateApiVersion(%+v) returned unexpected error: %s", req, err)
		}

	}
}

func BenchmarkUpdateApiSpec(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		//Update the first spec of the first api version of every api
		req := &rpc.UpdateApiSpecRequest{
			ApiSpec: &rpc.ApiSpec{
				Name:        fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, 1, 1),
				Description: fmt.Sprintf("Updated description of spec-%d for v%d of %s", 1, 1, apiId),
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"description"},
			},
		}
		b.StartTimer()
		_, err := client.UpdateApiSpec(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil {
			b.Errorf("UpdateApiSpec(%+v) returned unexpected error: %s", req, err)
		}
	}

}

func BenchmarkUpdateArtifact(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		//Read the artifact of the first spec of the first api version of every api
		messageMimeType := "text/plain"
		artifactId := "self-parent-link"
		messageContents := []byte(fmt.Sprintf("Updated - %s", artifactId))
		req := &rpc.ReplaceArtifactRequest{
			Artifact: &rpc.Artifact{
				Name:     fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d/artifacts/%s", rootResource, apiId, 1, 1, artifactId),
				MimeType: messageMimeType,
				Contents: messageContents,
			},
		}
		b.StartTimer()
		_, err = client.ReplaceArtifact(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil {
			b.Errorf("ReplaceArtifact(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkGetApi(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()
	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)

		req := &rpc.GetApiRequest{
			Name: fmt.Sprintf("%s/apis/%s", rootResource, apiId),
		}
		b.StartTimer()
		_, err := client.GetApi(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil { //If version does not exist
			b.Errorf("GetApi(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkGetApiVersion(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)
		// Read the first api version of every api
		versionName := fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, 1)
		req := &rpc.GetApiVersionRequest{
			Name: versionName,
		}
		b.StartTimer()
		_, err := client.GetApiVersion(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil { //If version does not exist
			b.Errorf("GetApiVersion(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkGetApiSpec(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)

		//Read the first spec of the first api revision of every api
		req := &rpc.GetApiSpecRequest{
			Name: fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, 1, 1),
		}
		b.StartTimer()
		_, err = client.GetApiSpec(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil {
			b.Errorf("GetApiSpec(%+v) returned unexpected error: %s", req, err)
		}
	}
}
func BenchmarkGetApiSpecContents(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)

		//Read the contents of the first spec of the first api version of every api
		req := &rpc.GetApiSpecContentsRequest{
			Name: fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, 1, 1),
		}
		// Only take reading for the first version
		b.StartTimer()
		_, err = client.GetApiSpecContents(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil {
			b.Errorf("GetApiSpecContents(%+v) returned unexpected error: %s", req, err)
		}
	}
}
func BenchmarkGetArtifact(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()
	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)

		//Read the artifact of the first spec of the first api version of every api
		artifactName := fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d/artifacts/self-parent-link", rootResource, apiId, 1, 1)
		req := &rpc.GetArtifactRequest{
			Name: artifactName,
		}
		b.StartTimer()
		_, err := client.GetArtifact(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil { //If version does not exist
			b.Errorf("GetArtifact(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkListApis_Pagination(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	pageToken := ""

	for i := 1; i <= b.N; i++ {
		req := &rpc.ListApisRequest{
			Parent:    rootResource,
			PageSize:  int32(pageSize),
			PageToken: pageToken,
		}
		b.StartTimer()
		obj, err := client.ListApis(ctx, req)
		b.StopTimer()
		pageToken = obj.NextPageToken
		/**
		Go Benchmarking always runs all the benchmark tests once before it runs
		for the duration specified. So the first row always returns error.
		*/
		if i > 1 && err != nil {
			b.Errorf("ListApis(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkListApis_Filter(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)
		req := &rpc.ListApisRequest{
			Parent:   rootResource,
			PageSize: int32(pageSize),
			Filter:   fmt.Sprintf("name.startsWith('%s/apis/%s')", rootResource, apiId),
		}
		b.StartTimer()
		_, err := client.ListApis(ctx, req)
		b.StopTimer()
		/**
		Go Benchmarking always runs all the benchmark tests once before it runs
		for the duration specified. So the first row always returns error.
		*/
		if i > 1 && err != nil {
			b.Errorf("ListApis(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkListApiVersions(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)

		req := &rpc.ListApiVersionsRequest{
			Parent: fmt.Sprintf("%s/apis/%s", rootResource, apiId),
		}
		b.StartTimer()
		_, err := client.ListApiVersions(ctx, req)
		b.StopTimer()
		/**
		Go Benchmarking always runs all the benchmark tests once before it runs
		for the duration specified. So the first row always returns error.
		*/
		if i > 1 && err != nil {
			b.Errorf("ListApiVersions(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkListApiSpecs(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)

		req := &rpc.ListApiSpecsRequest{
			Parent: fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, 1),
		}
		b.StartTimer()
		_, err := client.ListApiSpecs(ctx, req)
		b.StopTimer()
		/**
		Go Benchmarking always runs all the benchmark tests once before it runs
		for the duration specified. So the first row always returns error.
		*/
		if i > 1 && err != nil {
			b.Errorf("ListApiSpecs(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkListApiSpecArtifacts(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, b)

		req := &rpc.ListArtifactsRequest{
			Parent: fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, 1, 1),
		}
		b.StartTimer()
		_, err := client.ListArtifacts(ctx, req)
		b.StopTimer()
		/**
		Go Benchmarking always runs all the benchmark tests once before it runs
		for the duration specified. So the first row always returns error.
		*/
		if i > 1 && err != nil {
			b.Errorf("ListArtifacts(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkDeleteArtifact(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()
	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		for j := 1; j <= versionCount; j++ {
			artifactName := fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d/artifacts/self-parent-link", rootResource, apiId, j, j)
			req := &rpc.DeleteArtifactRequest{
				Name: artifactName,
			}
			// Only take reading for the first version
			if j == 1 {
				b.StartTimer()
			}
			_, err := client.DeleteArtifact(ctx, req)
			if j == 1 {
				b.StopTimer()
			}
			if i > 1 && err != nil {
				b.Errorf("DeleteArtifact(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkDeleteApiSpec(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		for j := 1; j <= versionCount; j++ {
			specName := fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, j, j)
			req := &rpc.DeleteApiSpecRequest{
				Name:  specName,
				Force: true,
			}
			// Only take reading for the first version
			if j == 1 {
				b.StartTimer()
			}
			_, err := client.DeleteApiSpec(ctx, req)
			// Only take reading for the first version
			if j == 1 {
				b.StopTimer()
			}
			if i > 1 && err != nil {
				b.Errorf("DeleteApiSpec(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkDeleteApiVersion(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()
	b.StopTimer()
	b.ResetTimer()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		for j := 1; j <= versionCount; j++ {
			versionName := fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, j)
			req := &rpc.DeleteApiVersionRequest{
				Name:  versionName,
				Force: true,
			}
			// Only take reading for the first version
			if j == 1 {
				b.StartTimer()
			}
			_, err = client.DeleteApiVersion(ctx, req)
			if j == 1 {
				b.StopTimer()
			}
			if i > 1 && err != nil {
				b.Errorf("DeleteApiVersion(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkDeleteApi(b *testing.B) {
	rootResource := getRootResource(b)

	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	client := gapicClient.GrpcClient()

	for i := 1; i <= b.N; i++ {
		pauseAfter1000Iterations(i, b)

		apiId := getApiName(i, b)

		apiName := fmt.Sprintf("%s/apis/%s", rootResource, apiId)
		req := &rpc.DeleteApiRequest{Name: apiName}
		b.StartTimer()
		_, err := client.DeleteApi(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil {
			b.Errorf("DeleteApi(%+v) returned unexpected error: %s", req, err)
		}
	}
}
