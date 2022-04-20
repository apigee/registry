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
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/genproto/protobuf/field_mask"
)

var (
	registryProject string
	rootResource    string
)

func init() {
	flag.StringVar(&registryProject, "registry_project", "bench", "Name of the Registry project")
	rootResource = fmt.Sprintf("projects/%s/locations/global", registryProject)
}
func getApiName(i int) string {
	return fmt.Sprintf("test-%d", i)
}

func createGrpcClient(b *testing.B) (error, rpc.RegistryClient, context.Context) {
	b.Helper()
	ctx := context.Background()
	gapicClient, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
		return err, nil, nil
	}
	client := gapicClient.GrpcClient()
	b.StopTimer()
	b.ResetTimer()
	return nil, client, ctx
}

func CreateApi(apiId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) *rpc.Api {
	b.Helper()
	req := &rpc.CreateApiRequest{
		Parent: rootResource,
		ApiId:  apiId,
		Api: &rpc.Api{
			DisplayName: apiId,
			Description: fmt.Sprintf("Description for %s", apiId),
		},
	}
	if record {
		b.StartTimer()
	}
	api, err := client.CreateApi(ctx, req)
	if record {
		b.StopTimer()
	}

	if apiId != getApiName(1) && err != nil {
		b.Errorf("CreateApi(%+v) returned unexpected error: %s", req, err)
	}
	return api
}
func UpdateApi(apiId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()
	apiName := fmt.Sprintf("%s/apis/%s", rootResource, apiId)
	req := &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:        apiName,
			DisplayName: fmt.Sprintf("Updated %s", apiId),
		}, UpdateMask: &field_mask.FieldMask{
			Paths: []string{"display_name"},
		},
	}

	if record {
		b.StartTimer()
	}
	_, err := client.UpdateApi(ctx, req)
	if record {
		b.StopTimer()
	}

	if err != nil {
		b.Errorf("UpdateApi(%+v) returned unexpected error: %s", req, err)
	}
}
func GetApi(apiId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()
	apiName := fmt.Sprintf("%s/apis/%s", rootResource, apiId)
	req := &rpc.GetApiRequest{Name: apiName}

	if record {
		b.StartTimer()
	}
	_, err := client.GetApi(ctx, req)
	if record {
		b.StopTimer()
	}

	if err != nil {
		b.Errorf("DeleteApi(%+v) returned unexpected error: %s", req, err)
	}
}
func DeleteApi(apiId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()
	apiName := fmt.Sprintf("%s/apis/%s", rootResource, apiId)
	req := &rpc.DeleteApiRequest{Name: apiName}

	if record {
		b.StartTimer()
	}
	_, err := client.DeleteApi(ctx, req)
	if record {
		b.StopTimer()
	}

	if apiId != getApiName(1) && err != nil {
		b.Errorf("DeleteApi(%+v) returned unexpected error: %s", req, err)
	}
}

func CreateApiVersion(apiId string, versionId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()
	req := &rpc.CreateApiVersionRequest{
		Parent: fmt.Sprintf("%s/apis/%s", rootResource, apiId),
		ApiVersion: &rpc.ApiVersion{
			DisplayName: versionId,
			Description: fmt.Sprintf("%s of %s", versionId, apiId),
		},
		ApiVersionId: versionId,
	}
	if record {
		b.StartTimer()
	}
	_, err := client.CreateApiVersion(ctx, req)
	if record {
		b.StopTimer()
	}
	if apiId != getApiName(1) && err != nil {
		b.Errorf("CreateApiVersion(%+v) returned unexpected error: %s", req, err)
	}
}
func UpdateApiVersion(apiId string, versionId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()
	versionName := fmt.Sprintf("%s/apis/%s/versions/%s", rootResource, apiId, versionId)

	req := &rpc.UpdateApiVersionRequest{
		ApiVersion: &rpc.ApiVersion{
			Name:        versionName,
			DisplayName: fmt.Sprintf("Updated %s", versionId),
			Description: fmt.Sprintf("Updated description %s of %s", versionId, apiId),
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"display_name", "description"},
		},
	}
	if record {
		b.StartTimer()
	}
	_, err := client.UpdateApiVersion(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("UpdateApiVersion(%+v) returned unexpected error: %s", req, err)
	}
}
func GetApiVersion(apiId string, versionId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()
	versionName := fmt.Sprintf("%s/apis/%s/versions/%s", rootResource, apiId, versionId)
	req := &rpc.GetApiVersionRequest{
		Name: versionName,
	}
	if record {
		b.StartTimer()
	}
	_, err := client.GetApiVersion(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("GetApiVersion(%+v) returned unexpected error: %s", req, err)
	}
}
func DeleteApiVersion(apiId string, versionId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()
	versionName := fmt.Sprintf("%s/apis/%s/versions/%s", rootResource, apiId, versionId)
	req := &rpc.DeleteApiVersionRequest{
		Name:  versionName,
		Force: true,
	}
	if record {
		b.StartTimer()
	}
	_, err := client.DeleteApiVersion(ctx, req)
	if record {
		b.StopTimer()
	}
	if apiId != getApiName(1) && err != nil {
		b.Errorf("DeleteApiVersion(%+v) returned unexpected error: %s", req, err)
	}
}

func CreateApiSpec(apiId string, versionId string, specId string, contents []byte, mimeType string, fileName string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	req := &rpc.CreateApiSpecRequest{
		Parent:    fmt.Sprintf("%s/apis/%s/versions/%s", rootResource, apiId, versionId),
		ApiSpecId: specId,
		ApiSpec: &rpc.ApiSpec{
			MimeType:    mimeType,
			Contents:    contents,
			Filename:    fileName,
			Description: fmt.Sprintf("Spec for api[%s] version[%s]", apiId, versionId),
		},
	}

	if record {
		b.StartTimer()
	}
	_, err := client.CreateApiSpec(ctx, req)
	if record {
		b.StopTimer()
	}
	if apiId != getApiName(1) && err != nil {
		b.Errorf("CreateApiSpec(%+v) returned unexpected error: %s", req, err)
	}
}
func UpdateApiSpec(apiId string, versionId string, specId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()
	specName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s", rootResource, apiId, versionId, specId)

	req := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:        specName,
			Description: fmt.Sprintf("Updated description %s of %s", specId, versionId),
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"description"},
		},
	}
	if record {
		b.StartTimer()
	}
	_, err := client.UpdateApiSpec(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("UpdateApiSpec(%+v) returned unexpected error: %s", req, err)
	}
}
func GetApiSpec(apiId string, versionId string, specId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	specName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s", rootResource, apiId, versionId, specId)

	req := &rpc.GetApiSpecRequest{
		Name: specName,
	}

	if record {
		b.StartTimer()
	}
	_, err := client.GetApiSpec(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("GetApiSpec(%+v) returned unexpected error: %s", req, err)
	}
}
func GetApiSpecContents(apiId string, versionId string, specId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	specName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s", rootResource, apiId, versionId, specId)

	req := &rpc.GetApiSpecContentsRequest{
		Name: specName,
	}

	if record {
		b.StartTimer()
	}
	_, err := client.GetApiSpecContents(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("GetApiSpecContents(%+v) returned unexpected error: %s", req, err)
	}
}
func DeleteApiSpec(apiId string, versionId string, specId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	specName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s", rootResource, apiId, versionId, specId)

	req := &rpc.DeleteApiSpecRequest{
		Name:  specName,
		Force: true,
	}

	if record {
		b.StartTimer()
	}
	_, err := client.DeleteApiSpec(ctx, req)
	if record {
		b.StopTimer()
	}
	if apiId != getApiName(1) && err != nil {
		b.Errorf("DeleteApiSpec(%+v) returned unexpected error: %s", req, err)
	}
}

func CreateApiSpecArtifact(apiId string, versionId string, specId string, artifactId string, artifactContents string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	specName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s", rootResource, apiId, versionId, specId)

	messageContents := []byte(artifactContents)
	messageMimeType := "text/plain"
	// Set the artifact.
	req := &rpc.CreateArtifactRequest{
		Parent:     specName,
		ArtifactId: artifactId,
		Artifact: &rpc.Artifact{
			MimeType: messageMimeType,
			Contents: messageContents,
		},
	}

	if record {
		b.StartTimer()
	}
	_, err := client.CreateArtifact(ctx, req)
	if record {
		b.StopTimer()
	}
	if apiId != getApiName(1) && err != nil {
		b.Errorf("CreateArtifact(%+v) returned unexpected error: %s", req, err)
	}
}
func UpdateApiSpecArtifact(apiId string, versionId string, specId string, artifactId string, artifactContents string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	artifactName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s/artifacts/%s", rootResource, apiId, versionId, specId, artifactId)
	messageContents := []byte(artifactContents)
	messageMimeType := "text/plain"

	req := &rpc.ReplaceArtifactRequest{
		Artifact: &rpc.Artifact{
			Name:     artifactName,
			MimeType: messageMimeType,
			Contents: messageContents,
		},
	}

	if record {
		b.StartTimer()
	}
	_, err := client.ReplaceArtifact(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("ReplaceArtifact(%+v) returned unexpected error: %s", req, err)
	}
}
func GetApiSpecArtifact(apiId string, versionId string, specId string, artifactId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	artifactName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s/artifacts/%s", rootResource, apiId, versionId, specId, artifactId)

	req := &rpc.GetArtifactRequest{
		Name: artifactName,
	}

	if record {
		b.StartTimer()
	}
	_, err := client.GetArtifact(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("GetArtifact(%+v) returned unexpected error: %s", req, err)
	}
}
func GetApiSpecArtifactContents(apiId string, versionId string, specId string, artifactId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	artifactName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s/artifacts/%s", rootResource, apiId, versionId, specId, artifactId)

	req := &rpc.GetArtifactContentsRequest{
		Name: artifactName,
	}

	if record {
		b.StartTimer()
	}
	_, err := client.GetArtifactContents(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("GetArtifactContents(%+v) returned unexpected error: %s", req, err)
	}
}
func DeleteApiSpecArtifact(apiId string, versionId string, specId string, artifactId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	b.Helper()

	artifactName := fmt.Sprintf("%s/apis/%s/versions/%s/specs/%s/artifacts/%s", rootResource, apiId, versionId, specId, artifactId)

	req := &rpc.DeleteArtifactRequest{
		Name: artifactName,
	}

	if record {
		b.StartTimer()
	}
	_, err := client.DeleteArtifact(ctx, req)
	if record {
		b.StopTimer()
	}
	if apiId != getApiName(1) && err != nil {
		b.Errorf("DeleteApiSpecArtifact(%+v) returned unexpected error: %s", req, err)
	}
}

func ListApis_FirstPageNoFilter(pageSize int, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	req := &rpc.ListApisRequest{
		Parent:   rootResource,
		PageSize: int32(pageSize),
	}
	if record {
		b.StartTimer()
	}
	_, err := client.ListApis(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("ListApis(%+v) returned unexpected error: %s", req, err)
	}
}
func ListApis_PaginationNoFilter(pageSize int, pageToken string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) string {
	req := &rpc.ListApisRequest{
		Parent:    rootResource,
		PageSize:  int32(pageSize),
		PageToken: pageToken,
	}
	if record {
		b.StartTimer()
	}
	list, err := client.ListApis(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("ListApis(%+v) returned unexpected error: %s", req, err)
		return ""
	}
	return list.NextPageToken
}
func ListApis_FirstPageFilter(pageSize int, apiId string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) {
	req := &rpc.ListApisRequest{
		Parent:   rootResource,
		PageSize: int32(pageSize),
		Filter:   fmt.Sprintf("name.startsWith('%s/apis/%s')", rootResource, apiId),
	}
	if record {
		b.StartTimer()
	}
	_, err := client.ListApis(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("ListApis(%+v) returned unexpected error: %s", req, err)
	}
}
func ListApis_PaginationFilter(pageSize int, apiId string, pageToken string, b *testing.B, record bool, client rpc.RegistryClient, ctx context.Context) string {
	req := &rpc.ListApisRequest{
		Parent:    rootResource,
		PageSize:  int32(pageSize),
		Filter:    fmt.Sprintf("name.startsWith('%s/apis/%s')", rootResource, apiId),
		PageToken: pageToken,
	}
	if record {
		b.StartTimer()
	}
	list, err := client.ListApis(ctx, req)
	if record {
		b.StopTimer()
	}
	if err != nil {
		b.Errorf("ListApis(%+v) returned unexpected error: %s", req, err)
		return ""
	}
	return list.NextPageToken
}

//
//func BenchmarkListApis_Pagination(b *testing.B) {
//
//	ctx := context.Background()
//	gapicClient, err := connection.NewClient(ctx)
//	if err != nil {
//		b.Fatalf("Setup: Failed to create client: %s", err)
//	}
//	client := gapicClient.GrpcClient()
//
//	pageToken := ""
//
//	for i := 1; i <= b.N; i++ {
//		req := &rpc.ListApisRequest{
//			Parent:    rootResource,
//			PageSize:  int32(pageSize),
//			PageToken: pageToken,
//		}
//		b.StartTimer()
//		obj, err := client.ListApis(ctx, req)
//		b.StopTimer()
//		pageToken = obj.NextPageToken
//		/**
//		Go Benchmarking always runs all the benchmark tests once before it runs
//		for the duration specified. So the first row always returns error.
//		*/
//		if i > 1 && err != nil {
//			b.Errorf("ListApis(%+v) returned unexpected error: %s", req, err)
//		}
//	}
//}
//
//func BenchmarkListApis_Filter(b *testing.B) {
//
//	ctx := context.Background()
//	gapicClient, err := connection.NewClient(ctx)
//	if err != nil {
//		b.Fatalf("Setup: Failed to create client: %s", err)
//	}
//	client := gapicClient.GrpcClient()
//
//	for i := 1; i <= b.N; i++ {
//		apiId := getApiName(i)
//		req := &rpc.ListApisRequest{
//			Parent:   rootResource,
//			PageSize: int32(pageSize),
//			Filter:   fmt.Sprintf("name.startsWith('%s/apis/%s')", rootResource, apiId),
//		}
//		b.StartTimer()
//		_, err := client.ListApis(ctx, req)
//		b.StopTimer()
//		/**
//		Go Benchmarking always runs all the benchmark tests once before it runs
//		for the duration specified. So the first row always returns error.
//		*/
//		if i > 1 && err != nil {
//			b.Errorf("ListApis(%+v) returned unexpected error: %s", req, err)
//		}
//	}
//}
//
//func BenchmarkListApiVersions(b *testing.B) {
//
//	ctx := context.Background()
//	gapicClient, err := connection.NewClient(ctx)
//	if err != nil {
//		b.Fatalf("Setup: Failed to create client: %s", err)
//	}
//	client := gapicClient.GrpcClient()
//
//	for i := 1; i <= b.N; i++ {
//		apiId := getApiName(i)
//
//		req := &rpc.ListApiVersionsRequest{
//			Parent: fmt.Sprintf("%s/apis/%s", rootResource, apiId),
//		}
//		b.StartTimer()
//		_, err := client.ListApiVersions(ctx, req)
//		b.StopTimer()
//		/**
//		Go Benchmarking always runs all the benchmark tests once before it runs
//		for the duration specified. So the first row always returns error.
//		*/
//		if i > 1 && err != nil {
//			b.Errorf("ListApiVersions(%+v) returned unexpected error: %s", req, err)
//		}
//	}
//}
//
//func BenchmarkListApiSpecs(b *testing.B) {
//
//	ctx := context.Background()
//	gapicClient, err := connection.NewClient(ctx)
//	if err != nil {
//		b.Fatalf("Setup: Failed to create client: %s", err)
//	}
//	client := gapicClient.GrpcClient()
//
//	for i := 1; i <= b.N; i++ {
//		apiId := getApiName(i)
//
//		req := &rpc.ListApiSpecsRequest{
//			Parent: fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, 1),
//		}
//		b.StartTimer()
//		_, err := client.ListApiSpecs(ctx, req)
//		b.StopTimer()
//		/**
//		Go Benchmarking always runs all the benchmark tests once before it runs
//		for the duration specified. So the first row always returns error.
//		*/
//		if i > 1 && err != nil {
//			b.Errorf("ListApiSpecs(%+v) returned unexpected error: %s", req, err)
//		}
//	}
//}
//
//func BenchmarkListApiSpecArtifacts(b *testing.B) {
//
//	ctx := context.Background()
//	gapicClient, err := connection.NewClient(ctx)
//	if err != nil {
//		b.Fatalf("Setup: Failed to create client: %s", err)
//	}
//	client := gapicClient.GrpcClient()
//
//	for i := 1; i <= b.N; i++ {
//		apiId := getApiName(i)
//
//		req := &rpc.ListArtifactsRequest{
//			Parent: fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, 1, 1),
//		}
//		b.StartTimer()
//		_, err := client.ListArtifacts(ctx, req)
//		b.StopTimer()
//		/**
//		Go Benchmarking always runs all the benchmark tests once before it runs
//		for the duration specified. So the first row always returns error.
//		*/
//		if i > 1 && err != nil {
//			b.Errorf("ListArtifacts(%+v) returned unexpected error: %s", req, err)
//		}
//	}
//}
