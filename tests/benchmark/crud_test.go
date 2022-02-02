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
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/genproto/protobuf/field_mask"
)

func uniqueID() string {
	uniqueId, valueExists := os.LookupEnv("API_NAME_PREFIX")
	if valueExists {
		return uniqueId
	}
	return "benchtest"
}

func getRootResource() string {
	registryProject, valueExists := os.LookupEnv("REGISTRY_PROJECT_IDENTIFIER")
	if valueExists {
		return fmt.Sprintf("projects/%s/locations/global", registryProject)
	}
	return "projects/bench/locations/global"
}

func getNumberOfVersions() int {
	strNumberOfVersions, exists := os.LookupEnv("REGISTRY_VERSION_COUNT")
	numberOfVersions, err := strconv.Atoi(strNumberOfVersions)
	if err != nil || !exists {
		numberOfVersions = 1
	}
	return numberOfVersions
}

func getApiName(i int, uniqueID string) string {
	return fmt.Sprintf("%s-%d", uniqueID, i)
}

var rootResource = getRootResource()

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

func BenchmarkCreateApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)
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
		*/
		if i > 1 && err != nil {
			b.Errorf("CreateApi(%+v) returned unexpected error: %s", req, err)
		}
	}
}

func BenchmarkCreateApiVersion(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)
		for j := 1; j <= getNumberOfVersions(); j++ {
			req := &rpc.CreateApiVersionRequest{
				Parent: fmt.Sprintf("%s/apis/%s", rootResource, apiId),
				ApiVersion: &rpc.ApiVersion{
					DisplayName: fmt.Sprintf("v%d", j),
					Description: fmt.Sprintf("v%d of %s", j, apiId),
				},
				ApiVersionId: fmt.Sprintf("v%d", j),
			}
			b.StartTimer()
			_, err := client.CreateApiVersion(ctx, req)
			b.StopTimer()
			if i > 1 && err != nil {
				b.Errorf("CreateApiVersion(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkCreateApiSpec(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {

			buf, err := readAndGZipFile(filepath.Join("testdata", "openapi1.yaml"))
			if err != nil {
				b.Errorf("BenchmarkCreateApiSpec could not read openapi1.yaml : %s", err)
			} else {
				req := &rpc.CreateApiSpecRequest{
					Parent: fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, j),
					ApiSpec: &rpc.ApiSpec{
						MimeType:    "application/x.openapi+gzip;version=3.0.0",
						Contents:    buf.Bytes(),
						Filename:    fmt.Sprintf("openapi-%s-v%d-spec.yaml", apiId, j),
						Description: fmt.Sprintf("Spec for api[%s] version[v%d]", apiId, j),
					},
					ApiSpecId: fmt.Sprintf("spec-%d", j),
				}
				b.StartTimer()
				_, err = client.CreateApiSpec(ctx, req)
				b.StopTimer()
				if i > 1 && err != nil {
					b.Errorf("CreateApiSpec(%+v) returned unexpected error: %s", req, err)
				}
			}
		}
	}
}

func BenchmarkCreateArtifact(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {
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
			b.StartTimer()
			_, err = client.CreateArtifact(ctx, req)
			b.StopTimer()
			if i > 1 && err != nil {
				b.Errorf("CreateArtifact(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkUpdateApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)
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
			b.StartTimer()
			_, err = client.UpdateApi(ctx, updateReq)
			b.StopTimer()
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
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)
		for j := 1; j <= getNumberOfVersions(); j++ {
			req := &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, j),
					DisplayName: fmt.Sprintf("Updated v%d", j),
					Description: fmt.Sprintf("Updated description v%d of %s", j, apiId),
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
}

func BenchmarkUpdateApiSpecVersion(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)
		for j := 1; j <= getNumberOfVersions(); j++ {
			req := &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name:        fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, j, j),
					Description: fmt.Sprintf("Updated description of spec-%d for v%d of %s", j, j, apiId),
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
}

func BenchmarkUpdateArtifact(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {
			messageMimeType := "text/plain"
			artifactId := "self-parent-link"
			messageContents := []byte(fmt.Sprintf("Updated - %s", artifactId))
			req := &rpc.ReplaceArtifactRequest{
				Artifact: &rpc.Artifact{
					Name:     fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d/artifacts/%s", rootResource, apiId, j, j, artifactId),
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
}

func BenchmarkGetApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)
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
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {
			versionName := fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, j)
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
}

func BenchmarkGetApiSpec(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {

			req := &rpc.GetApiSpecRequest{
				Name: fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, j, j),
			}
			b.StartTimer()
			_, err = client.GetApiSpec(ctx, req)
			b.StopTimer()
			if i > 1 && err != nil {
				b.Errorf("GetApiSpec(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}
func BenchmarkGetApiSpecContents(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {

			req := &rpc.GetApiSpecContentsRequest{
				Name: fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, j, j),
			}
			b.StartTimer()
			_, err = client.GetApiSpecContents(ctx, req)
			b.StopTimer()
			if i > 1 && err != nil {
				b.Errorf("GetApiSpecContents(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}
func BenchmarkGetArtifact(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {
			artifactName := fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d/artifacts/self-parent-link", rootResource, apiId, j, j)
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
}

func BenchmarkDeleteArtifact(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {
			artifactName := fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d/artifacts/self-parent-link", rootResource, apiId, j, j)
			req := &rpc.DeleteArtifactRequest{
				Name: artifactName,
			}
			b.StartTimer()
			err := client.DeleteArtifact(ctx, req)
			b.StopTimer()
			if i > 1 && err != nil {
				b.Errorf("DeleteArtifact(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkDeleteApiSpec(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)

		for j := 1; j <= getNumberOfVersions(); j++ {
			specName := fmt.Sprintf("%s/apis/%s/versions/v%d/specs/spec-%d", rootResource, apiId, j, j)
			req := &rpc.DeleteApiSpecRequest{
				Name:  specName,
				Force: true,
			}
			b.StartTimer()
			err := client.DeleteApiSpec(ctx, req)
			b.StopTimer()
			if i > 1 && err != nil {
				b.Errorf("DeleteApiSpec(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkDeleteApiVersion(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	b.StopTimer()
	b.ResetTimer()
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)
		for j := 1; j <= getNumberOfVersions(); j++ {
			versionName := fmt.Sprintf("%s/apis/%s/versions/v%d", rootResource, apiId, j)
			req := &rpc.DeleteApiVersionRequest{
				Name:  versionName,
				Force: true,
			}
			b.StartTimer()
			err = client.DeleteApiVersion(ctx, req)
			b.StopTimer()
			if i > 1 && err != nil {
				b.Errorf("DeleteApiVersion(%+v) returned unexpected error: %s", req, err)
			}
		}
	}
}

func BenchmarkDeleteApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}
	apiNamePrefix := uniqueID()
	for i := 1; i <= b.N; i++ {
		apiId := getApiName(i, apiNamePrefix)
		apiName := fmt.Sprintf("%s/apis/%s", rootResource, apiId)
		req := &rpc.DeleteApiRequest{Name: apiName}
		b.StartTimer()
		err = client.DeleteApi(ctx, req)
		b.StopTimer()
		if i > 1 && err != nil {
			b.Errorf("CreateApi(%+v) returned unexpected error: %s", req, err)
		}

	}
}
