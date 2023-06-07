// Copyright 2022 Google LLC.
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

package apply

import (
	"context"
	"os"
	"testing"

	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"gopkg.in/yaml.v3"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

const sampleDir = "testdata/sample"

func TestApply(t *testing.T) {
	project := names.Project{ProjectID: "apply-test"}
	parent := project.String() + "/locations/global"

	ctx := context.Background()
	grpctest.SetupRegistry(ctx, t, project.ProjectID, nil)

	// Test various normal invocations of `registry apply`
	tests := []struct {
		desc string
		args []string
	}{
		{
			desc: "apis-registry.yaml",
			args: []string{"-f", sampleDir + "/apis/registry.yaml", "--parent", parent, "--jobs", "1"},
		},
		{
			desc: "artifacts-lifecycle.yaml",
			args: []string{"-f", sampleDir + "/artifacts/lifecycle.yaml", "--parent", parent, "--jobs", "1"},
		},
		{
			desc: "sample",
			args: []string{"-f", sampleDir, "-R", "--parent", parent, "--jobs", "1"},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cmd := Command()
			cmd.SetArgs(test.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
			}
		})
	}
}

func TestApplyErrors(t *testing.T) {
	project := names.Project{ProjectID: "apply-test-errors"}
	parent := project.String() + "/locations/global"

	ctx := context.Background()
	grpctest.SetupRegistry(ctx, t, project.ProjectID, nil)

	// clear the configured registry.project
	config, err := connection.ActiveConfig()
	if err != nil {
		t.Fatalf("Setup: Failed to get registry configuration: %s", err)
	}
	config.Project = ""
	connection.SetConfig(config)

	// Test various erroneous invocations of `registry apply`
	tests := []struct {
		desc string
		args []string
	}{
		{
			desc: "input file not found",
			args: []string{"-f", sampleDir + "/missing.yaml", "--parent", parent},
		},
		{
			desc: "no input file specified",
			args: []string{"--parent", parent},
		},
		{
			desc: "no parent specified",
			args: []string{"-f", sampleDir + "/apis/registry.yaml"},
		},
		{
			desc: "invalid parent specified",
			args: []string{"-f", sampleDir + "/apis/registry.yaml", "--parent", "projects/invalid/locations/global"},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cmd := Command()
			cmd.SetArgs(test.args)
			if err := cmd.Execute(); err == nil {
				t.Fatalf("Execute() with args %+v succeeded, expected error", cmd.Args)
			}
		})
	}
}

func TestApply_Stdin(t *testing.T) {
	project := names.Project{ProjectID: "apply-test-stdin"}
	parent := project.String() + "/locations/global"

	ctx := context.Background()
	grpctest.SetupRegistry(ctx, t, project.ProjectID, nil)

	tests := []struct {
		desc string
		file string
		args []string
	}{
		{
			desc: "apis-registry.yaml",
			file: sampleDir + "/apis/registry.yaml",
			args: []string{"-f", "-", "--parent", parent, "--jobs", "1"},
		},
		{
			desc: "artifacts-lifecycle.yaml",
			file: sampleDir + "/artifacts/lifecycle.yaml",
			args: []string{"-f", "-", "--parent", parent, "--jobs", "1"},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			r, err := os.Open(test.file)
			if err != nil {
				t.Fatalf("Setup: failed to read file: %s", err)
			}
			defer r.Close()

			cmd := Command()
			cmd.SetArgs(test.args)
			cmd.SetIn(r)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %+v returned error: %s", test.args, err)
			}
		})
	}
}

func TestArtifactStorage(t *testing.T) {
	project := names.Project{ProjectID: "apply-test"}
	parent := project.String() + "/locations/global"
	artifactName, _ := names.ParseArtifact(parent + "/artifacts/lifecycle")
	file := sampleDir + "/artifacts/lifecycle.yaml"

	ctx := context.Background()
	client, _ := grpctest.SetupRegistry(ctx, t, project.ProjectID, nil)

	// store as yaml
	t.Run("store as yaml", func(t *testing.T) {
		cmd := Command()
		args := []string{"-f", file, "--parent", parent, "--jobs", "1", "--yaml"}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", args, err)
		}

		if err := visitor.GetArtifact(ctx, client, artifactName, true, func(ctx context.Context, message *rpc.Artifact) error {
			var encArt encoding.Artifact
			artBytes, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if err := yaml.Unmarshal(artBytes, &encArt); err != nil {
				t.Fatal(err)
			}

			artYamlBytes, err := yaml.Marshal(encArt.Data)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(artYamlBytes, message.GetContents(), nil); diff != "" {
				t.Errorf("unexpected diff (-want +got):\n%s", diff)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("store as proto", func(t *testing.T) {
		cmd := Command()
		args := []string{"-f", file, "--parent", parent, "--jobs", "1"}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", args, err)
		}

		ac, err := client.GetArtifactContents(ctx, &rpc.GetArtifactContentsRequest{
			Name: artifactName.String(),
		})
		if err != nil {
			t.Fatal(err)
		}
		lifecycle := new(apihub.Lifecycle)
		if err = proto.Unmarshal(ac.Data, lifecycle); err != nil {
			t.Fatal(err)
		}

		want := &apihub.Lifecycle{
			Id:          "lifecycle", // deprecated field
			Kind:        "Lifecycle", // deprecated field
			DisplayName: "Lifecycle",
			Description: "A series of stages that an API typically moves through in its lifetime",
			Stages: []*apihub.Lifecycle_Stage{
				{
					Id:           "concept",
					DisplayName:  "Concept",
					Description:  "Description of the business case and user needs for why an API should exist",
					DisplayOrder: 0,
				},
				{
					Id:           "design",
					DisplayName:  "Design",
					Description:  "Definition of the interface details and proposal of the API contract",
					DisplayOrder: 1,
				},
				{
					Id:           "develop",
					DisplayName:  "Develop",
					Description:  "Implementation of the service and its API",
					DisplayOrder: 2,
				},
				{
					Id:           "preview",
					DisplayName:  "Preview",
					Description:  "Staging of implementations in the pre-production phase",
					DisplayOrder: 3,
				},
				{
					Id:           "production",
					DisplayName:  "Production",
					Description:  "API available for production workloads",
					DisplayOrder: 4,
				},
				{
					Id:           "deprecated",
					DisplayName:  "Deprecated",
					Description:  "API not recommended for new consumers",
					DisplayOrder: 5,
				},
				{
					Id:           "retired",
					DisplayName:  "Retired",
					Description:  "API no longer available for use",
					DisplayOrder: 6,
				},
			},
		}

		opts := cmp.Options{
			protocmp.Transform(),
		}
		if !cmp.Equal(want, lifecycle, opts) {
			t.Errorf("unexpected diff (-want +got):\n%s", cmp.Diff(want, lifecycle, opts))
		}
	})
}
