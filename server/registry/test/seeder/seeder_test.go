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

package seeder

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

type fakeServer struct {
	*registry.RegistryServer
	Resources []string
}

func (s *fakeServer) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	s.Resources = append(s.Resources, fmt.Sprintf("projects/%s", req.GetProjectId()))
	return nil, nil
}

func (s *fakeServer) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	s.Resources = append(s.Resources, req.Api.GetName())
	return nil, nil
}

func (s *fakeServer) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	s.Resources = append(s.Resources, req.ApiVersion.GetName())
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

func (s *fakeServer) UpdateApiDeployment(ctx context.Context, req *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	s.Resources = append(s.Resources, req.ApiDeployment.GetName())
	return &rpc.ApiDeployment{
		Name:       req.ApiDeployment.GetName(),
		RevisionId: fmt.Sprintf("%.8s", uuid.New().String()),
	}, nil
}

func (s *fakeServer) TagApiDeploymentRevision(ctx context.Context, req *rpc.TagApiDeploymentRevisionRequest) (*rpc.ApiDeployment, error) {
	tagged := fmt.Sprintf("%s@%s", strings.Split(req.GetName(), "@")[0], req.GetTag())
	s.Resources = append(s.Resources, tagged)
	return nil, nil
}

func (s *fakeServer) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	s.Resources = append(s.Resources, fmt.Sprintf("%s/artifacts/%s", req.GetParent(), req.GetArtifactId()))
	return nil, nil
}

func (s *fakeServer) ReplaceArtifact(ctx context.Context, req *rpc.ReplaceArtifactRequest) (*rpc.Artifact, error) {
	s.Resources = append(s.Resources, req.Artifact.GetName())
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
				&rpc.ApiDeployment{Name: "projects/p/locations/global/apis/a/deployments/d"},
				&rpc.Artifact{Name: "projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a"},
			},
			want: []string{
				"projects/p",
				"projects/p/locations/global/apis/a",
				"projects/p/locations/global/apis/a/deployments/d",
				"projects/p/locations/global/apis/a/versions/v",
				"projects/p/locations/global/apis/a/versions/v/specs/s",
				"projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a",
			},
		},
		{
			desc: "resources can be created implicitly",
			seed: []RegistryResource{
				&rpc.Artifact{Name: "projects/p/locations/global/apis/a/versions/v/specs/s/artifacts/a"},
				&rpc.Artifact{Name: "projects/p/locations/global/apis/a/deployments/d/artifacts/a"},
			},
			want: []string{
				"projects/p",
				"projects/p/locations/global/apis/a",
				"projects/p/locations/global/apis/a/deployments/d",
				"projects/p/locations/global/apis/a/deployments/d/artifacts/a",
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
			desc: "deployment revisions can be created when attributes change",
			seed: []RegistryResource{
				&rpc.ApiDeployment{
					Name:            "projects/p/locations/global/apis/a/deployments/d",
					ApiSpecRevision: "first",
				},
				&rpc.ApiDeployment{
					Name:            "projects/p/locations/global/apis/a/deployments/d",
					ApiSpecRevision: "second",
				},
			},
			want: []string{
				"projects/p",
				"projects/p/locations/global/apis/a",
				"projects/p/locations/global/apis/a/deployments/d",
				"projects/p/locations/global/apis/a/deployments/d",
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

			sortStrings := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if diff := cmp.Diff(test.want, server.Resources, sortStrings); diff != "" {
				t.Errorf("SeedRegistry(%v) performed unexpected resource creation sequence (-want +got):\n%s", test.seed, diff)
			}
		})
	}
}
