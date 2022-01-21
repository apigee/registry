// Copyright 2021 Google LLC
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

package controller

import (
	"os"
	"context"
	"testing"
	"fmt"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/server/registry/names"
	"github.com/apigee/registry/rpc"
)

const testProject = "local-tasks-test"
const specName = "projects/local-tasks-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml"

// This is a fake process which executes the commands passed in test cases
func TestProcessHelper(t *testing.T) {
	
	core.FakeTestProcess(t, func (cmd string, args []string) {
		switch cmd {
		case "registry":
			subCmd := fmt.Sprintf("%s %s", args[0], args[1])
			if subCmd == "compute lint" {
				fmt.Fprint(os.Stderr, "compute doesn't have any child command lint")
				os.Exit(2)
			} else if subCmd == "compute conformance" {
				fmt.Printf("Computing conformance report %s/artifacts/conformance-openapi", specName)
				os.Exit(0)
			} else if subCmd == "compute complexity" {
				fmt.Printf("Computing complexity %s/artifacts/complexity", specName)
				os.Exit(0)
			} else if subCmd == "compute summary" {
				fmt.Printf("Computing summary projects/%s/locations/global/artifacts/summary", testProject)
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
			os.Exit(2)
		case "swagger-codegen":
			fmt.Printf("Generating code...")
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
			os.Exit(2)
		}
	})
}

func TestRun(t *testing.T) {
	tests := []struct {
		desc string
		task *ExecCommandTask
		wantErr bool 
	}{
		{
			desc: "Success",
			task: &ExecCommandTask{
				Action: &Action{Command: fmt.Sprintf("registry compute complexity %s", specName)},
				CreateCommand: core.GetFakeCommandGenerator("TestProcessHelper"),
			},
		},
		{
			desc: "Failure",
			task: &ExecCommandTask{
				Action: &Action{Command: fmt.Sprintf("registry compute lint %s", specName)},
				CreateCommand: core.GetFakeCommandGenerator("TestProcessHelper"),
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			err := test.task.Run(ctx)
			if test.wantErr {
				if err == nil {
					t.Errorf("Run() was successful, wanted error")
				} 
			} else {
				if err != nil {
					t.Errorf("Unexpected error from Run(): %s", err)
				}
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {

	tests := []struct {
		desc           string
		task *ExecCommandTask
		wantErr bool
	}{
		{
			desc: "Success",
			task: &ExecCommandTask{
				Action: &Action{
					Command: fmt.Sprintf("registry compute complexity %s", specName),
				},
			},
		},
		{
			desc: "Error",
			task: &ExecCommandTask{
				Action: &Action{
					Command: fmt.Sprintf("registry compute lint %s", specName),
				},
				// CreateCommand: core.GetFakeCommandGenerator("TestProcessHelper"),
			},
			wantErr: true,
		},
		{
			desc: "Resolve Command",
			task: &ExecCommandTask{
				Action: &Action{
					Command: "registry resolve projects/demo/artifacts/test-manifest",
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			logger := log.FromContext(ctx)

			//Attach the fake command generator to the task
			test.task.CreateCommand = core.GetFakeCommandGenerator("TestProcessHelper")
			err := test.task.ExcecuteCommand(ctx, logger)
			if test.wantErr {
				if err == nil {
					t.Errorf("ExecuteCommand() was succcessful, wanted error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error from ExecuteCommand(): %s", err)
				}
			}
		})
	}
}

func TestExecuteCommandWithReceipt(t *testing.T) {

	tests := []struct {
		desc           string
		task *ExecCommandTask
		wantErr bool
		wantReceipt bool
	}{
		// Test: Both command execution and receipt creation is successful, overall success.
		{
			desc: "Receipt Success",
			task: &ExecCommandTask{
				Action: &Action{
					Command: fmt.Sprintf("registry compute summary projects/%s", testProject),
					GeneratedResource: fmt.Sprintf("projects/%s/locations/global/artifacts/summary", testProject),
					RequiresReceipt: true,
				},
			},
			wantReceipt: true,
		},
		// Test: Command execution is successful, receipt creation fails, overall error.
		{
			desc: "Receipt Failure",
			task: &ExecCommandTask{
				Action: &Action{
					Command: fmt.Sprintf("registry compute conformance %s", specName),
					GeneratedResource: fmt.Sprintf("%s/artifacts/conformance-openapi",  specName),
					RequiresReceipt: true,
				},
			},
			wantErr: true,
			wantReceipt: false,
		},
		// Test: Command execution fails, hence receipt should not be created, overall error.
		{
			desc: "Command Failure",
			task:  &ExecCommandTask{
				Action: &Action{
					Command: fmt.Sprintf("registry compute score projects/%s",  testProject),
					GeneratedResource:  fmt.Sprintf("projects/%s/locations/global/artifacts/score", testProject),
					RequiresReceipt: true,
				},
			},
			wantErr: true,
			wantReceipt: false,
		},
		{
			desc: "3p Command",
			task:  &ExecCommandTask{
				Action: &Action{
					Command: fmt.Sprintf("swagger-codegen generate -i projects/%s/locatios/global/apis/-/versions/-/specs/- -l csharp",  testProject),
					GeneratedResource:  fmt.Sprintf("projects/%s/locations/global/artifacts/swagger-codegen", testProject),
					RequiresReceipt: true,
				},
			},
			wantErr: false,
			wantReceipt: true,	
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			client, err := connection.NewClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}

			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}

			// Setup a project for successful creation of project level summary artifacts
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + testProject,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}

			_, err = adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
				ProjectId: testProject,
				Project: &rpc.Project{
					DisplayName: "Demo",
					Description: "A demo catalog",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create project %s: %s", testProject, err.Error())
			}

			//Attach the fake command generator to the task
			test.task.CreateCommand = core.GetFakeCommandGenerator("TestProcessHelper")
			
			logger := log.FromContext(ctx)
			err = test.task.ExcecuteCommand(ctx, logger)

			// Check if error is returned
			if test.wantErr {
				if err == nil {
					t.Errorf("ExecuteCommand() was successful, wanted error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error from ExecuteCommand(): %s", err)
				}
			}

			// Check if the artifact is created
			pattern := test.task.Action.GeneratedResource
			artifactName, err := names.ParseArtifact(pattern)
			if err != nil {
				t.Fatalf("Invalid artifact pattern: %s", pattern)
			}
			_, err = core.GetArtifact(ctx, client, artifactName, true, nil)

			if test.wantReceipt {
				if err != nil {
					t.Errorf("Expected ExecuteCommand() to create receipt artifact, artifact not found")
				}
			} else {
				if err == nil {
					t.Errorf("ExecuteCommand() should not create receipt artifact, artifact found")
				}
			}

			// Delete the demo project
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + testProject,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}

		})
	}

}

func TestTouchArtifact(t *testing.T){
	tests := []struct {
		desc    string
		action string
		artifactName string
		wantErr bool
	}{
		{
			desc: "Normal case",
			action: "registry test action",
			artifactName: fmt.Sprintf("projects/%s/locations/global/artifacts/summary", testProject),
			wantErr:  false,
		},
		{
			desc: "SetArtifact Error",
			action: fmt.Sprintf("registry compute complexity %s", specName),
			artifactName: fmt.Sprintf("%s/artifacts/complexity", specName),
			wantErr: true,
		},
		{
			desc: "ProtoMarshal Error",
			action: "\xa0\xa1", //invalid UTF-8 string for proto.Marshal to fail
			artifactName: fmt.Sprintf("projects/%s/locations/global/artifacts/summary", testProject),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}

			// Setup a project for successful creation of project level summary artifacts
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + testProject,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}

			_, err = adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
				ProjectId: testProject,
				Project: &rpc.Project{
					DisplayName: "Demo",
					Description: "A demo catalog",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create project %s: %s", testProject, err.Error())
			}

			task := &ExecCommandTask{}

			err = task.touchArtifact(ctx, test.artifactName, test.action)
			if test.wantErr{
				if err == nil {
					t.Errorf("touchArtifact() succeeded, expected error")
				}
			} else {
				if err != nil {
					t.Errorf("touchArtifact() generated unexpected error: %s", err)
				}
			}

			// Delete the demo project
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + testProject,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}
		})
	}
}