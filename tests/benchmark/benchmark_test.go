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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func BenchmarkApi(b *testing.B) {
	err, client, ctx := createGrpcClient(b)
	if err != nil {
		return
	}
	operations := [5]string{"Create", "Update", "Get", "List_pagesize_20", "Delete"}

	for _, operation := range operations {
		b.Run(fmt.Sprintf("%s", operation), func(b *testing.B) {
			for i := 1; i <= b.N; i++ {
				apiId := getApiName(i)
				switch operation {
				case "Create":
					CreateApi(apiId, b, true, client, ctx)
					break
				case "Update":
					UpdateApi(apiId, b, true, client, ctx)
					break
				case "Get":
					GetApi(apiId, b, true, client, ctx)
					break
				case "List_pagesize_20":
					ListApis_FirstPageNoFilter(20, b, true, client, ctx)
					break
				case "Delete":
					DeleteApi(apiId, b, true, client, ctx)
					break
				}
			}
		})
	}
}

func BenchmarkApiVersion(b *testing.B) {
	err, client, ctx := createGrpcClient(b)
	if err != nil {
		return
	}
	versionId := "v1"
	operations := [4]string{"Create", "Update", "Get", "Delete"}
	for _, operation := range operations {
		b.Run(fmt.Sprintf("%s", operation), func(b *testing.B) {
			for i := 1; i <= b.N; i++ {
				apiId := getApiName(i)
				switch operation {
				case "Create":
					CreateApi(apiId, b, false, client, ctx)
					CreateApiVersion(apiId, versionId, b, true, client, ctx)
					break
				case "Update":
					UpdateApiVersion(apiId, versionId, b, true, client, ctx)
					break
				case "Get":
					GetApiVersion(apiId, versionId, b, true, client, ctx)
					break
				case "Delete":
					DeleteApiVersion(apiId, versionId, b, true, client, ctx)
					DeleteApi(apiId, b, false, client, ctx)
					break
				}
			}
		})
	}
}

func BenchmarkApiSpec(b *testing.B) {
	err, client, ctx := createGrpcClient(b)
	if err != nil {
		return
	}
	fileName := "openapi.yaml"
	specPath := filepath.Join("testdata", fileName)
	mimeType := "application/x.openapi;version=3.0.0"
	contents, err := ioutil.ReadFile(specPath)
	operations := [5]string{"Create", "Update", "Get", "GetContents", "Delete"}
	versionId := "v1"
	specId := "spec1"
	for _, operation := range operations {
		b.Run(fmt.Sprintf("%s", operation), func(b *testing.B) {
			for i := 1; i <= b.N; i++ {
				apiId := getApiName(i)
				switch operation {
				case "Create":
					CreateApi(apiId, b, false, client, ctx)
					CreateApiVersion(apiId, versionId, b, false, client, ctx)
					CreateApiSpec(apiId, versionId, specId, contents, mimeType, fileName, b, true, client, ctx)
					break
				case "Update":
					UpdateApiSpec(apiId, versionId, specId, b, true, client, ctx)
					break
				case "Get":
					GetApiSpec(apiId, versionId, specId, b, true, client, ctx)
					break
				case "GetContents":
					GetApiSpecContents(apiId, versionId, specId, b, true, client, ctx)
					break
				case "Delete":
					DeleteApiSpec(apiId, versionId, specId, b, true, client, ctx)
					DeleteApiVersion(apiId, versionId, b, false, client, ctx)
					DeleteApi(apiId, b, false, client, ctx)
					break
				}
			}
		})
	}
}

func BenchmarkApiSpecArtifact(b *testing.B) {
	err, client, ctx := createGrpcClient(b)
	if err != nil {
		return
	}

	fileName := "openapi.yaml"
	specPath := filepath.Join("testdata", fileName)
	mimeType := "application/x.openapi;version=3.0.0"
	contents, err := ioutil.ReadFile(specPath)
	versionId := "v1"
	specId := "spec1"
	artifactId := "artifact-1"

	operations := [5]string{"Create", "Update", "Get", "GetContents", "Delete"}
	for _, operation := range operations {
		b.Run(fmt.Sprintf("%s", operation), func(b *testing.B) {
			for i := 1; i <= b.N; i++ {
				apiId := getApiName(i)
				switch operation {
				case "Create":
					CreateApi(apiId, b, false, client, ctx)
					CreateApiVersion(apiId, versionId, b, false, client, ctx)
					CreateApiSpec(apiId, versionId, specId, contents, mimeType, fileName, b, false, client, ctx)
					CreateApiSpecArtifact(apiId, versionId, specId, artifactId, "artifact-value", b, true, client, ctx)
					break
				case "Update":
					UpdateApiSpecArtifact(apiId, versionId, specId, artifactId, "updated-artifact-value", b, true, client, ctx)
					break
				case "Get":
					GetApiSpecArtifact(apiId, versionId, specId, artifactId, b, true, client, ctx)
					break
				case "GetContents":
					GetApiSpecArtifactContents(apiId, versionId, specId, artifactId, b, true, client, ctx)
					break
				case "Delete":
					DeleteApiSpecArtifact(apiId, versionId, specId, artifactId, b, true, client, ctx)
					DeleteApiSpec(apiId, versionId, specId, b, false, client, ctx)
					DeleteApiVersion(apiId, versionId, b, false, client, ctx)
					DeleteApi(apiId, b, false, client, ctx)
					break
				}
			}
		})
	}
}

func BenchmarkListApis(b *testing.B) {
	err, client, ctx := createGrpcClient(b)
	//versionCount := 3
	if err != nil {
		return
	}
	fileName := "openapi.yaml"
	specPath := filepath.Join("testdata", fileName)
	mimeType := "application/x.openapi;version=3.0.0"
	contents, err := ioutil.ReadFile(specPath)
	setup := false

	numberOfApis := [3]int{100, 500, 1000}
	versionCounts := [2]int{1, 3}
	operations := [4]string{"FirstPageNoFilter", "PaginationNoFilter", "FirstPageFilter", "PaginationFilter"}
	pageSizes := [3]int{20, 50, 100}
	for _, apiCount := range numberOfApis {
		for _, versionCount := range versionCounts {
			for _, pageSize := range pageSizes {
				for _, operation := range operations {
					b.Run(fmt.Sprintf("Apis(%d)/versions(%d)/%s/PageSize(%d)]", apiCount, versionCount, operation, pageSize), func(b *testing.B) {
						if !setup {
							for i := 1; i <= apiCount; i++ {
								apiId := getApiName(i)
								CreateApi(apiId, b, false, client, ctx)
								for j := 1; j <= versionCount; j++ {
									versionId := fmt.Sprintf("v%d", j)
									specId := fmt.Sprintf("spec-%d", j)
									artifactId := fmt.Sprintf("artifact-%d", j)
									CreateApiVersion(apiId, versionId, b, false, client, ctx)
									CreateApiSpec(apiId, versionId, specId, contents, mimeType, fileName, b, false, client, ctx)
									CreateApiSpecArtifact(apiId, versionId, specId, artifactId, "artifact-value", b, false, client, ctx)
								}
							}
							setup = true
						}
						pageToken := ""
						for i := 1; i <= b.N; i++ {
							switch operation {
							case "FirstPageNoFilter":
								ListApis_FirstPageNoFilter(pageSize, b, true, client, ctx)
								break
							case "PaginationNoFilter":
								pageToken = ListApis_PaginationNoFilter(pageSize, pageToken, b, true, client, ctx)
								break
							case "FirstPageFilter":
								ListApis_FirstPageFilter(pageSize, "api1", b, true, client, ctx)
								break
							case "PaginationFilter":
								pageToken = ListApis_PaginationFilter(pageSize, "api1", pageToken, b, true, client, ctx)
								break
							}
						}
						if operation == operations[len(operations)-1] && pageSize == pageSizes[len(pageSizes)-1] {
							//Since this is the last benchmark combination we will
							//cleanup the test data
							for i := 1; i <= apiCount; i++ {
								apiId := getApiName(i)
								for j := 1; j <= versionCount; j++ {
									versionId := fmt.Sprintf("v%d", j)
									specId := fmt.Sprintf("spec-%d", j)
									artifactId := fmt.Sprintf("artifact-%d", j)
									DeleteApiSpecArtifact(apiId, versionId, specId, artifactId, b, false, client, ctx)
									DeleteApiSpec(apiId, versionId, specId, b, false, client, ctx)
									DeleteApiVersion(apiId, versionId, b, false, client, ctx)
								}
								DeleteApi(apiId, b, false, client, ctx)
							}
						}
					})
				}
			}
		}
	}
}
