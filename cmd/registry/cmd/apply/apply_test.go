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

package apply

import (
	"context"
	"os"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/server/registry"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
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
