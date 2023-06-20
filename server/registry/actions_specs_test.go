// Copyright 2020 Google LLC.
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

package registry

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var (
	// Example spec contents for an OpenAPI JSON spec.
	specContents = []byte(`{"openapi": "3.0.0", "info": {"title": "My API", "version": "v1"}, "paths": {}}`)
)

func sha256hash(bytes []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(bytes))
}

func TestCreateApiSpec(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.CreateApiSpecRequest
		want *rpc.ApiSpec
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "my-spec",
				ApiSpec: &rpc.ApiSpec{
					Filename:    "openapi.json",
					Description: "My Description",
					MimeType:    "application/x.openapi;version=3.0.0",
					SourceUri:   "https://www.example.com/openapi.json",
					Contents:    specContents,
					Labels: map[string]string{
						"label-key": "label-value",
					},
					Annotations: map[string]string{
						"annotation-key": "annotation-value",
					},
				},
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Filename:    "openapi.json",
				Description: "My Description",
				MimeType:    "application/x.openapi;version=3.0.0",
				SizeBytes:   int32(len(specContents)),
				Hash:        sha256hash(specContents),
				SourceUri:   "https://www.example.com/openapi.json",
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
		},
		{
			desc: "empty resource",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "my-spec",
				ApiSpec:   nil,
			},
			want: &rpc.ApiSpec{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			created, err := server.CreateApiSpec(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateApiSpec(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateApiSpec(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if created.RevisionId == "" {
				t.Errorf("CreateApiSpec(%+v) returned unexpected revision_id %q, expected non-empty ID", test.req, created.GetRevisionId())
			}

			if created.CreateTime == nil || created.RevisionCreateTime == nil || created.RevisionUpdateTime == nil {
				t.Errorf("CreateApiSpec(%+v) returned unset create_time (%v), revision_create_time (%v), or revision_update_time (%v)", test.req, created.CreateTime, created.RevisionCreateTime, created.RevisionUpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.RevisionCreateTime.AsTime()) {
				t.Errorf("CreateApiSpec(%+v) returned unexpected timestamps: create_time %v != revision_create_time %v", test.req, created.CreateTime, created.RevisionCreateTime)
			} else if !created.RevisionCreateTime.AsTime().Equal(created.RevisionUpdateTime.AsTime()) {
				t.Errorf("CreateApiSpec(%+v) returned unexpected timestamps: revision_create_time %v != revision_update_time %v", test.req, created.RevisionCreateTime, created.RevisionUpdateTime)
			}

			t.Run("GetApiSpec", func(t *testing.T) {
				req := &rpc.GetApiSpecRequest{
					Name: created.GetName(),
				}

				got, err := server.GetApiSpec(ctx, req)
				if err != nil {
					t.Fatalf("GetApiSpec(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(created, got, opts) {
					t.Errorf("GetApiSpec(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(created, got, opts))
				}
			})
		})
	}
}

func TestCreateApiSpecResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.CreateApiSpecRequest
		want codes.Code
	}{
		{
			desc: "parent not found",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v2",
				ApiSpecId: "valid-id",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "valid-id",
				ApiSpec:   nil,
			},
			want: codes.OK,
		},
		{
			desc: "missing custom identifier",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "specific revision",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "my-spec@12345678",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "this-identifier-is-invalid-because-it-exceeds-the-eighty-character-maximum-length",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "underscore_identifier",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen prefix",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api",
				ApiSpecId: "-identifier",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen suffix",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api",
				ApiSpecId: "identifier-",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier mixed case",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "IDentifier",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid contents for gzip mime type",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "my-spec",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip",
					Contents: []byte("these contents are not gzipped"),
				},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.CreateApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateApiSpecDuplicates(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.CreateApiSpecRequest
		want codes.Code
	}{
		{
			desc: "case sensitive",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
				ApiSpecId: "my-spec",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.AlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.CreateApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestGetApiSpec(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.GetApiSpecRequest
		want *rpc.ApiSpec
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Filename:    "openapi.json",
				Description: "My API Spec",
				MimeType:    "application/x.openapi;version=3.0.0",
				SourceUri:   "https://www.example.com/openapi.json",
				Contents:    specContents,
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
			req: &rpc.GetApiSpecRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Filename:    "openapi.json",
				Description: "My API Spec",
				MimeType:    "application/x.openapi;version=3.0.0",
				SizeBytes:   int32(len(specContents)),
				Hash:        sha256hash(specContents),
				SourceUri:   "https://www.example.com/openapi.json",
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.GetApiSpec(ctx, test.req)
			if err != nil {
				t.Fatalf("GetApiSpec(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("GetApiSpec(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}
		})
	}
}

func TestGetApiSpecResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.GetApiSpecRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.GetApiSpecRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "invalid name",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.GetApiSpecRequest{
				Name: "invalid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "revision not found",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.GetApiSpecRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec@missing",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.GetApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestGetApiSpecContents(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.GetApiSpecContentsRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "revision not found",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec@missing",
			},
			want: codes.NotFound,
		},
		{
			desc: "inappropriate contents suffix in resource name",
			seed: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Contents: []byte{},
			},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec/contents",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "gzip mimetype with empty contents",
			seed: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				MimeType: "application/x.openapi+gzip;version=3.0.0",
				Contents: []byte{},
			},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
			want: codes.FailedPrecondition,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.GetApiSpecContents(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetApiSpecContents(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func gZippedBytes(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = zw.Write(input)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestGetApiSpecContentsData(t *testing.T) {
	specContentsCompressed, err := gZippedBytes(specContents)
	if err != nil {
		t.Fatalf("Setup/Seeding: Failed to gzip sample spec %s", err)
	}
	tests := []struct {
		desc     string
		seed     *rpc.ApiSpec
		req      *rpc.GetApiSpecContentsRequest
		accept   string
		wantCode codes.Code
		wantData []byte
	}{
		{
			desc: "uncompressed mimetype with no accept-encoding header",
			seed: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				MimeType: "application/x.openapi;version=3.0.0",
				Contents: specContents,
			},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
			wantCode: codes.OK,
			wantData: specContents, // uncompressed upload, uncompressed response
		},
		{
			desc: "uncompressed mimetype with accept-encoding=gzip",
			seed: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				MimeType: "application/x.openapi;version=3.0.0",
				Contents: specContents,
			},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
			accept:   "gzip",
			wantCode: codes.OK,
			wantData: specContents, // we only return compressed data when it was uploaded by the client
		},
		{
			desc: "gzip mimetype with no accept-encoding header",
			seed: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				MimeType: "application/x.openapi+gzip;version=3.0.0",
				Contents: specContentsCompressed,
			},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
			wantCode: codes.OK,
			wantData: specContents, // compressed upload, uncompressed response
		},
		{
			desc: "gzip mimetype with accept-encoding=gzip",
			seed: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				MimeType: "application/x.openapi+gzip;version=3.0.0",
				Contents: specContentsCompressed,
			},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
			accept:   "gzip",
			wantCode: codes.OK,
			wantData: specContentsCompressed, // compressed upload, compressed response
		},
		{
			desc: "gzip mimetype with accept-encoding=some-unknown-type",
			seed: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				MimeType: "application/x.openapi+gzip;version=3.0.0",
				Contents: specContentsCompressed,
			},
			req: &rpc.GetApiSpecContentsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
			accept:   "some-unknown-type",
			wantCode: codes.OK,
			wantData: specContents, // uncompressed response because we don't support the requested encoding
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}
			if test.accept != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("accept-encoding", test.accept))
			}
			resp, err := server.GetApiSpecContents(ctx, test.req)
			if status.Code(err) != test.wantCode {
				t.Errorf("GetApiSpecContents(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.wantCode, err)
			}
			if !bytes.Equal(resp.Data, test.wantData) {
				t.Errorf("GetApiSpecContents(%+v) returned contents %v, want %v: %v", test.req, resp.Data, test.wantData, err)
			}
		})
	}
}

func TestListApiSpecs(t *testing.T) {
	tests := []struct {
		admin     bool
		desc      string
		seed      []*rpc.ApiSpec
		req       *rpc.ListApiSpecsRequest
		want      *rpc.ListApiSpecsResponse
		wantToken bool
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/specs/spec1"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3"},
				},
			},
		},
		{
			desc: "with specs containing multiple revisions",
			seed: []*rpc.ApiSpec{
				{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
				},
				{
					Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
					Contents: []byte(specContents),
				},
				{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
				},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:      "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
						Hash:      sha256hash(specContents),
						SizeBytes: int32(len(specContents)),
					},
					{
						Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
					},
				},
			},
		},
		{
			admin: true,
			desc:  "across all versions in a specific project and api",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/specs/my-spec"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/-",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/specs/my-spec"},
				},
			},
		},
		{
			admin: true,
			desc:  "across all apis and versions in a specific project",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v2/specs/my-spec"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/-/versions/-",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/my-project/locations/global/apis/other-api/versions/v2/specs/my-spec"},
				},
			},
		},
		{
			admin: true,
			desc:  "across all projects, apis, and versions",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/other-project/locations/global/apis/other-api/versions/v2/specs/my-spec"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/-/locations/global/apis/-/versions/-",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/other-project/locations/global/apis/other-api/versions/v2/specs/my-spec"},
				},
			},
		},
		{
			admin: true,
			desc:  "in a specific api and version across all projects",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v1/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/specs/my-spec"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/-/locations/global/apis/my-api/versions/v1",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				},
			},
		},
		{
			admin: true,
			desc:  "in a specific version across all projects and apis",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/other-project/locations/global/apis/other-api/versions/v1/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/specs/my-spec"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/-/locations/global/apis/-/versions/v1",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/other-project/locations/global/apis/other-api/versions/v1/specs/my-spec"},
				},
			},
		},
		{
			admin: true,
			desc:  "in all versions of a specific api across all projects",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v2/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v1/specs/my-spec"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/-/locations/global/apis/my-api/versions/-",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/other-project/locations/global/apis/my-api/versions/v2/specs/my-spec"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent:   "projects/my-project/locations/global/apis/my-api/versions/v1",
				PageSize: 1,
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{},
				},
			},
			wantToken: true,
			// Ordering is not guaranteed by API, so any resource may be returned.
			extraOpts: protocmp.IgnoreFields(new(rpc.ApiSpec), "name"),
		},
		{
			desc: "name equality filtering",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
				Filter: "name == 'projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2'",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.ApiSpec{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
					Description: "First ApiSpec",
				},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
				Filter: "description != ''",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
						Description: "First ApiSpec",
					},
				},
			},
		},
		{
			desc: "ordered by description",
			seed: []*rpc.ApiSpec{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
					Description: "111: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
					Description: "333: this should be returned third",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3",
					Description: "222: this should be returned second",
				},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api/versions/v1",
				OrderBy: "description",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
						Description: "111: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3",
						Description: "222: this should be returned second",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
						Description: "333: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by description descending",
			seed: []*rpc.ApiSpec{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
					Description: "111: this should be returned third",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
					Description: "333: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3",
					Description: "222: this should be returned second",
				},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api/versions/v1",
				OrderBy: "description desc",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
						Description: "333: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3",
						Description: "222: this should be returned second",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
						Description: "111: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by description then by name",
			seed: []*rpc.ApiSpec{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
					Description: "222: this should be returned second or third (the name is the tie-breaker)",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3",
					Description: "111: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
					Description: "222: this should be returned second or third (the name is the tie-breaker)",
				},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api/versions/v1",
				OrderBy: "description,name",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3",
						Description: "111: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1",
						Description: "222: this should be returned second or third (the name is the tie-breaker)",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
						Description: "222: this should be returned second or third (the name is the tie-breaker)",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.admin && adminServiceUnavailable() {
				t.Skip(testRequiresAdminService)
			}
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed...); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.ListApiSpecs(ctx, test.req)
			if err != nil {
				t.Fatalf("ListApiSpecs(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListApiSpecsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListApiSpecs(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListApiSpecs(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				t.Errorf("ListApiSpecs(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListApiSpecsResponseCodes(t *testing.T) {
	tests := []struct {
		admin bool
		desc  string
		seed  *rpc.ApiSpec
		req   *rpc.ListApiSpecsRequest
		want  codes.Code
	}{
		{
			desc: "parent version not found",
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent api not found",
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/-",
			},
			want: codes.NotFound,
		},
		{
			admin: true,
			desc:  "parent project not found",
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/locations/global/apis/-/versions/-",
			},
			want: codes.NotFound,
		},
		{
			desc: "negative page size",
			req: &rpc.ListApiSpecsRequest{
				PageSize: -1,
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid filter",
			req: &rpc.ListApiSpecsRequest{
				Filter: "this filter is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid page token",
			req: &rpc.ListApiSpecsRequest{
				PageToken: "this token is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by unknown field",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.ListApiSpecsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api/versions/v1",
				OrderBy: "something",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by private field",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.ListApiSpecsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api/versions/v1",
				OrderBy: "key",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering direction",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.ListApiSpecsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api/versions/v1",
				OrderBy: "description asc",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering format",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.ListApiSpecsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api/versions/v1",
				OrderBy: "description,",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.admin && adminServiceUnavailable() {
				t.Skip(testRequiresAdminService)
			}
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.ListApiSpecs(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ListApiSpecs(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApiSpecsSequence(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := []*rpc.ApiSpec{
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1"},
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2"},
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3"},
	}
	if err := seeder.SeedSpecs(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	listed := make([]*rpc.ApiSpec, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListApiSpecsRequest{
			Parent:   "projects/my-project/locations/global/apis/my-api/versions/v1",
			PageSize: 1,
		}

		got, err := server.ListApiSpecs(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecs(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiSpecs()); count != 1 {
			t.Errorf("ListApiSpecs(%+v) returned %d specs, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListApiSpecs(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.ApiSpecs...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test intermediate page after failure on first page")
	}

	t.Run("intermediate page", func(t *testing.T) {
		req := &rpc.ListApiSpecsRequest{
			Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiSpecs(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecs(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiSpecs()); count != 1 {
			t.Errorf("ListApiSpecs(%+v) returned %d specs, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListApiSpecs(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.ApiSpecs...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on intermediate page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListApiSpecsRequest{
			Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiSpecs(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecs(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiSpecs()); count != 1 {
			t.Errorf("ListApiSpecs(%+v) returned %d specs, expected exactly one", req, count)
		}

		if got.GetNextPageToken() != "" {
			t.Errorf("ListApiSpecs(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.ApiSpecs...)
	})

	if t.Failed() {
		t.Fatal("Cannot test sequence result after failure on final page")
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.ApiSpec), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
		cmpopts.SortSlices(func(a, b *rpc.ApiSpec) bool {
			return a.GetName() < b.GetName()
		}),
	}

	if !cmp.Equal(seed, listed, opts) {
		t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(seed, listed, opts))
	}
}

func TestListApiSpecsLargeCollection(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := make([]*rpc.ApiSpec, 0, 1001)
	for i := 1; i <= cap(seed); i++ {
		seed = append(seed, &rpc.ApiSpec{
			Name: fmt.Sprintf("projects/my-project/locations/global/apis/my-api/versions/v1/specs/s%03d", i),
		})
	}

	if err := seeder.SeedSpecs(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	// This test prevents the list sequence from ending before a known filter match is listed.
	// For simplicity, it does not guarantee the resource is returned on a later page.
	t.Run("filter", func(t *testing.T) {
		req := &rpc.ListApiSpecsRequest{
			Parent:   "projects/my-project/locations/global/apis/my-api/versions/v1",
			PageSize: 1,
			Filter:   "name == 'projects/my-project/locations/global/apis/my-api/versions/v1/specs/s099'",
		}

		got, err := server.ListApiSpecs(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecs(%+v) returned error: %s", req, err)
		}

		if len(got.GetApiSpecs()) == 1 && got.GetNextPageToken() != "" {
			t.Errorf("ListApiSpecs(%+v) returned a page token when the only matching resource has been listed: %+v", req, got)
		} else if len(got.GetApiSpecs()) == 0 && got.GetNextPageToken() == "" {
			t.Errorf("ListApiSpecs(%+v) returned an empty next page token before listing the only matching resource", req)
		} else if count := len(got.GetApiSpecs()); count > 1 {
			t.Errorf("ListApiSpecs(%+v) returned %d projects, expected at most one: %+v", req, count, got.GetApiSpecs())
		}
	})

	t.Run("max page size", func(t *testing.T) {
		req := &rpc.ListApiSpecsRequest{
			Parent:   "projects/my-project/locations/global/apis/my-api/versions/v1",
			PageSize: 1001,
		}

		got, err := server.ListApiSpecs(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecs(%+v) returned error: %s", req, err)
		}

		if len(got.GetApiSpecs()) != 1000 {
			t.Errorf("GetApiSpecs(%+v) should have returned 1000 items, got: %+v", req, len(got.GetApiSpecs()))
		} else if got.GetNextPageToken() == "" {
			t.Errorf("GetApiSpecs(%+v) should return a next page token", req)
		}
	})
}

func TestUpdateApiSpec(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.UpdateApiSpecRequest
		want *rpc.ApiSpec
	}{
		{
			desc: "allow missing updates existing resources",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My ApiSpec",
				Filename:    "openapi.json",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					Description: "My Updated ApiSpec",
				},
				UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"description"}},
				AllowMissing: true,
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My Updated ApiSpec",
				Filename:    "openapi.json",
			},
		},
		{
			desc: "allow missing creates missing resources",
			seed: &rpc.ApiSpec{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/sibling-spec",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				},
				AllowMissing: true,
			},
			want: &rpc.ApiSpec{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
		},
		{
			desc: "implicit nil mask",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My ApiSpec",
				Filename:    "openapi.json",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					Description: "My Updated ApiSpec",
				},
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My Updated ApiSpec",
				Filename:    "openapi.json",
			},
		},
		{
			desc: "implicit empty mask",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My ApiSpec",
				Filename:    "openapi.json",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					Description: "My Updated ApiSpec",
				},
				UpdateMask: &fieldmaskpb.FieldMask{},
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My Updated ApiSpec",
				Filename:    "openapi.json",
			},
		},
		{
			desc: "field specific mask",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My ApiSpec",
				Filename:    "openapi.json",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					Description: "My Updated ApiSpec",
					Filename:    "Ignored",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"description"}},
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My Updated ApiSpec",
				Filename:    "openapi.json",
			},
		},
		{
			desc: "full replacement wildcard mask",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My ApiSpec",
				Filename:    "openapi.json",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					Description: "My Updated ApiSpec",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				Description: "My Updated ApiSpec",
				Filename:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			updated, err := server.UpdateApiSpec(ctx, test.req)
			if err != nil {
				t.Fatalf("UpdateApiSpec(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}

			if !cmp.Equal(test.want, updated, opts) {
				t.Errorf("UpdateApiSpec(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, updated, opts))
			}

			t.Run("GetApiSpec", func(t *testing.T) {
				req := &rpc.GetApiSpecRequest{
					Name: updated.GetName(),
				}

				got, err := server.GetApiSpec(ctx, req)
				if err != nil {
					t.Fatalf("GetApiSpec(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(updated, got, opts) {
					t.Errorf("GetApiSpec(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(updated, got, opts))
				}
			})
		})
	}
}

func TestUpdateApiSpecResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.UpdateApiSpecRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "specific revision",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec@12345678",
				},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource body",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req:  &rpc.UpdateApiSpecRequest{},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource name",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "nonexistent field in mask",
			seed: &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"this field does not exist"}},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid contents for gzip mime type",
			seed: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				MimeType: "application/x.openapi",
				Contents: []byte("these contents are not gzipped"),
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					MimeType: "application/x.openapi+gzip",
					Contents: []byte("these contents are not gzipped"),
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"contents", "mime_type"}},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.UpdateApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestUpdateApiSpecSequence(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.UpdateApiSpecRequest
		want codes.Code
	}{
		{
			desc: "create using update with allow_missing=false",
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s",
				},
				AllowMissing: false,
			},
			want: codes.NotFound,
		},
		{
			desc: "create using update with allow_missing=true",
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s",
				},
				AllowMissing: true,
			},
			want: codes.OK,
		},
		{
			desc: "update existing resource with allow_missing=true",
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s",
				},
				AllowMissing: true,
			},
			want: codes.OK,
		},
		{
			desc: "update existing resource with allow_missing=false",
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s",
				},
				AllowMissing: false,
			},
			want: codes.OK,
		},
	}
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/a/versions/v"}
	if err := seeder.SeedVersions(ctx, server, seed); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	var createTime time.Time
	var revisionCreateTime time.Time
	var revisionUpdateTime time.Time
	// NOTE: in the following sequence of tests, each test depends on its predecessor.
	// Resources are successively created and updated using the "Update" RPC and the
	// tests verify that CreateTime/UpdateTime fields are modified appropriately.
	for i, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var result *rpc.ApiSpec
			var err error
			if result, err = server.UpdateApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
			if result != nil {
				if i == 1 {
					createTime = result.CreateTime.AsTime()
					revisionCreateTime = result.RevisionCreateTime.AsTime()
					revisionUpdateTime = result.RevisionUpdateTime.AsTime()
				} else {
					if !createTime.Equal(result.CreateTime.AsTime()) {
						t.Errorf("UpdateApiSpec create time changed after update (%v %v)", createTime, result.CreateTime.AsTime())
					}
					if !revisionCreateTime.Equal(result.RevisionCreateTime.AsTime()) {
						t.Errorf("UpdateApiSpec revision create time changed after update (%v %v)", revisionCreateTime, result.RevisionCreateTime.AsTime())
					}
					if !revisionUpdateTime.Before(result.RevisionUpdateTime.AsTime()) {
						t.Errorf("UpdateApiSpec update time did not increase after update (%v %v)", revisionUpdateTime, result.RevisionUpdateTime.AsTime())
					}
					revisionUpdateTime = result.RevisionUpdateTime.AsTime()
				}
			}
		})
	}
}

func TestDeleteApiSpec(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.DeleteApiSpecRequest
	}{
		{
			desc: "existing resource",
			seed: &rpc.ApiSpec{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
			req: &rpc.DeleteApiSpecRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.DeleteApiSpec(ctx, test.req); err != nil {
				t.Fatalf("DeleteApiSpec(%+v) returned error: %s", test.req, err)
			}

			t.Run("GetApiSpec", func(t *testing.T) {
				req := &rpc.GetApiSpecRequest{
					Name: test.req.GetName(),
				}

				if _, err := server.GetApiSpec(ctx, req); status.Code(err) != codes.NotFound {
					t.Fatalf("GetApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), codes.NotFound, err)
				}
			})
		})
	}
}

func TestDeleteApiSpecResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.DeleteApiSpecRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.DeleteApiSpecRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "specific revision",
			req: &rpc.DeleteApiSpecRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec@12345678",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "resource has children",
			seed: &rpc.Artifact{
				Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec/artifacts/my-artifact",
			},
			req: &rpc.DeleteApiSpecRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec",
			},
			want: codes.FailedPrecondition,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.DeleteApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteApiSpecCascading(t *testing.T) {
	var (
		ctx      = context.Background()
		server   = defaultTestServer(t)
		artifact = &rpc.Artifact{
			Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec/artifacts/my-artifact",
		}
	)

	if err := seeder.SeedArtifacts(ctx, server, artifact); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	req := &rpc.DeleteApiSpecRequest{
		Name:  "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec",
		Force: true,
	}

	if _, err := server.DeleteApiSpec(ctx, req); err != nil {
		t.Errorf("DeleteApiSpec(%+v) returned error: %s", req, err)
	}

	if _, err := server.GetApiSpec(ctx, &rpc.GetApiSpecRequest{Name: req.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetApiSpec(%q) returned status code %q, want %q: %s", req.GetName(), status.Code(err), codes.NotFound, err)
	}

	if _, err := server.GetArtifact(ctx, &rpc.GetArtifactRequest{Name: artifact.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetArtifact(%q) returned status code %q, want %q: %s", artifact.GetName(), status.Code(err), codes.NotFound, err)
	}
}
