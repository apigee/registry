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

package seeder

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

type fakeServer struct {
	*server.RegistryServer
	Resources []string
}

func (s *fakeServer) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	s.Resources = append(s.Resources, fmt.Sprintf("projects/%s", req.GetProjectId()))
	return nil, nil
}

func (s *fakeServer) CreateApi(ctx context.Context, req *rpc.CreateApiRequest) (*rpc.Api, error) {
	s.Resources = append(s.Resources, fmt.Sprintf("%s/apis/%s", req.GetParent(), req.GetApiId()))
	return nil, nil
}

func (s *fakeServer) CreateApiVersion(ctx context.Context, req *rpc.CreateApiVersionRequest) (*rpc.ApiVersion, error) {
	s.Resources = append(s.Resources, fmt.Sprintf("%s/versions/%s", req.GetParent(), req.GetApiVersionId()))
	return nil, nil
}

func (s *fakeServer) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	s.Resources = append(s.Resources, req.ApiSpec.GetName())
	return &rpc.ApiSpec{
		Name:       req.ApiSpec.GetName(),
		RevisionId: fmt.Sprintf("%.8s", uuid.New().String()),
	}, nil
}

func (s *fakeServer) TagApiSpecRevision(ctx context.Context, req *rpc.TagApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	tagged := fmt.Sprintf("%s@%s", strings.Split(req.GetName(), "@")[0], req.GetTag())
	s.Resources = append(s.Resources, tagged)
	return nil, nil
}

func (s *fakeServer) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	s.Resources = append(s.Resources, fmt.Sprintf("%s/artifacts/%s", req.GetParent(), req.GetArtifactId()))
	return nil, nil
}

func TestSeedRegistry(t *testing.T) {
	tests := []struct {
		desc string
		seed []RegistryResource
		want []string
	}{
		{
			desc: "resources can be created explicitly",
			seed: []RegistryResource{
				&rpc.Project{Name: "projects/p"},
				&rpc.Api{Name: "projects/p/locations/global/apis/a"},
				&rpc.ApiVersion{Name: "projects/p/locations/global/apis/a/versions/v"},
				&rpc.ApiSpec{Name: "projects/p/locations/global/apis/a/versions/v/specs/s"},
				&rpc.Artifact{Name: "projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a"},
			},
			want: []string{
				"projects/p",
				"projects/p/locations/global/apis/a",
				"projects/p/locations/global/apis/a/versions/v",
				"projects/p/locations/global/apis/a/versions/v/specs/s",
				"projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a",
			},
		},
		{
			desc: "resources can be created implicitly",
			seed: []RegistryResource{
				&rpc.Artifact{Name: "projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a"},
			},
			want: []string{
				"projects/p",
				"projects/p/locations/global/apis/a",
				"projects/p/locations/global/apis/a/versions/v",
				"projects/p/locations/global/apis/a/versions/v/specs/s",
				"projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a",
			},
		},
		{
			desc: "resources can be created out of order",
			seed: []RegistryResource{
				&rpc.Artifact{Name: "projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a"},
				&rpc.ApiSpec{Name: "projects/p/locations/global/apis/a/versions/v/specs/s"},
				&rpc.ApiVersion{Name: "projects/p/locations/global/apis/a/versions/v"},
				&rpc.Api{Name: "projects/p/locations/global/apis/a"},
				&rpc.Project{Name: "projects/p"},
			},
			want: []string{
				"projects/p",
				"projects/p/locations/global/apis/a",
				"projects/p/locations/global/apis/a/versions/v",
				"projects/p/locations/global/apis/a/versions/v/specs/s",
				"projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a",
			},
		},
		{
			desc: "specs revisions can be created when contents change",
			seed: []RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/p/locations/global/apis/a/versions/v/specs/s",
					Contents: []byte("first"),
				},
				&rpc.ApiSpec{
					Name:     "projects/p/locations/global/apis/a/versions/v/specs/s",
					Contents: []byte("first"),
				},
				&rpc.ApiSpec{
					Name:     "projects/p/locations/global/apis/a/versions/v/specs/s",
					Contents: []byte("second"),
				},
			},
			want: []string{
				"projects/p",
				"projects/p/locations/global/apis/a",
				"projects/p/locations/global/apis/a/versions/v",
				"projects/p/locations/global/apis/a/versions/v/specs/s",
				"projects/p/locations/global/apis/a/versions/v/specs/s",
			},
		},
		{
			desc: "spec revision tags can be created",
			seed: []RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/p/locations/global/apis/a/versions/v/specs/s",
					RevisionTags: []string{
						"first-tag",
						"second-tag",
					},
				},
			},
			want: []string{
				"projects/p",
				"projects/p/locations/global/apis/a",
				"projects/p/locations/global/apis/a/versions/v",
				"projects/p/locations/global/apis/a/versions/v/specs/s",
				"projects/p/locations/global/apis/a/versions/v/specs/s@first-tag",
				"projects/p/locations/global/apis/a/versions/v/specs/s@second-tag",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var (
				ctx    = context.Background()
				server = new(fakeServer)
			)

			if err := SeedRegistry(ctx, server, test.seed...); err != nil {
				t.Errorf("SeedRegistry(%v) returned error: %s", test.seed, err)
			}

			if diff := cmp.Diff(test.want, server.Resources); diff != "" {
				t.Errorf("SeedRegistry(%v) performed unexpected resource creation sequence (-want +got):\n%s", test.seed, diff)
			}
		})
	}
}

func TestSeedRegistry_Errors(t *testing.T) {
	tests := []struct {
		desc string
		seed []RegistryResource
	}{
		{
			desc: "duplicate projects",
			seed: []RegistryResource{
				&rpc.Project{Name: "projects/p"},
				&rpc.Project{Name: "projects/p"},
			},
		},
		{
			desc: "duplicate apis",
			seed: []RegistryResource{
				&rpc.Api{Name: "projects/p/locations/global/apis/a"},
				&rpc.Api{Name: "projects/p/locations/global/apis/a"},
			},
		},
		{
			desc: "duplicate versions",
			seed: []RegistryResource{
				&rpc.ApiVersion{Name: "projects/p/locations/global/apis/a/versions/v"},
				&rpc.ApiVersion{Name: "projects/p/locations/global/apis/a/versions/v"},
			},
		},
		{
			desc: "duplicate artifacts",
			seed: []RegistryResource{
				&rpc.Artifact{Name: "projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a"},
				&rpc.Artifact{Name: "projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var (
				ctx    = context.Background()
				server = new(fakeServer)
			)

			if err := SeedRegistry(ctx, server, test.seed...); err == nil {
				t.Errorf("SeedRegistry(%v) returned without error", test.seed)
			}
		})
	}
}
