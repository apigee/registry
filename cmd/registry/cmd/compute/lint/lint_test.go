// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lint

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/apply"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestLint(t *testing.T) {
	command := Command()
	command.SilenceErrors = true
	command.SilenceUsage = true
	if err := command.Execute(); err == nil {
		t.Fatalf("Execute() with no args succeeded and should have failed")
	}
}

func TestInvalidComputeLint(t *testing.T) {
	tests := []struct {
		desc string
		args []string
	}{
		{
			desc: "no-linter-specified",
			args: []string{"spec"},
		},
		{
			desc: "missing-linter-specified",
			args: []string{"spec", "--linter", "nonexistent"},
		},
		{
			desc: "empty-linter-specified",
			args: []string{"spec", "--linter", ""},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			command := Command()
			command.SilenceErrors = true
			command.SilenceUsage = true
			command.SetArgs(test.args)
			if err := command.Execute(); err == nil {
				t.Fatalf("Execute() with no args succeeded and should have failed")
			}
		})
	}
}

func TestComputeLint(t *testing.T) {
	project := names.Project{ProjectID: "lint-test"}
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, project.ProjectID, nil)

	config, err := connection.ActiveConfig()
	if err != nil {
		t.Fatalf("Setup: Failed to get registry configuration: %s", err)
	}
	config.Project = project.ProjectID
	connection.SetConfig(config)

	applyCmd := apply.Command()
	applyCmd.SetArgs([]string{"-f", "../complexity/testdata/apigeeregistry", "-R"})
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("Failed to apply test API")
	}

	t.Run("protos", func(t *testing.T) {
		specName := project.Api("apigeeregistry").Version("v1").Spec("protos")
		lintCmd := Command()
		lintCmd.SetArgs([]string{specName.String(), "--linter", "test"})
		if err := lintCmd.Execute(); err != nil {
			t.Fatalf("Compute lint failed: %s", err)
		}
		artifactName := specName.Artifact("lint-test")
		if err = visitor.GetArtifact(ctx, registryClient, artifactName, true, func(ctx context.Context, message *rpc.Artifact) error {
			var lint style.Lint
			if err = patch.UnmarshalContents(message.Contents, message.MimeType, &lint); err != nil {
				return err
			}
			return nil
		}); err != nil {
			t.Fatalf("Error getting artifact: %s", err)
		}
	})

	t.Run("openapi", func(t *testing.T) {
		specName := project.Api("apigeeregistry").Version("v1").Spec("openapi")
		lintCmd := Command()
		lintCmd.SetArgs([]string{specName.String(), "--linter", "test"})
		if err := lintCmd.Execute(); err != nil {
			t.Fatalf("Compute lint failed: %s", err)
		}
		artifactName := specName.Artifact("lint-test")
		if err = visitor.GetArtifact(ctx, registryClient, artifactName, true, func(ctx context.Context, message *rpc.Artifact) error {
			var lint style.Lint
			if err = patch.UnmarshalContents(message.Contents, message.MimeType, &lint); err != nil {
				return err
			}
			return nil
		}); err != nil {
			t.Fatalf("Error getting artifact: %s", err)
		}
	})

	t.Run("discovery", func(t *testing.T) {
		specName := project.Api("apigeeregistry").Version("v1").Spec("discovery")
		lintCmd := Command()
		lintCmd.SetArgs([]string{specName.String(), "--linter", "test"})
		if err := lintCmd.Execute(); err != nil {
			t.Fatalf("Compute lint failed: %s", err)
		}
		artifactName := specName.Artifact("lint-test")
		if err = visitor.GetArtifact(ctx, registryClient, artifactName, true, func(ctx context.Context, message *rpc.Artifact) error {
			var lint style.Lint
			if err = patch.UnmarshalContents(message.Contents, message.MimeType, &lint); err != nil {
				return err
			}
			return nil
		}); err != nil {
			t.Fatalf("Error getting artifact: %s", err)
		}
	})
}
