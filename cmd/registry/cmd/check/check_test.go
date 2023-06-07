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

package check

import (
	"bufio"
	"bytes"
	"context"
	"testing"

	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"gopkg.in/yaml.v3"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestCheck(t *testing.T) {
	ctx := context.Background()
	grpctest.SetupRegistry(ctx, t, "my-project", []seeder.RegistryResource{
		&rpc.ApiSpec{
			Name:     "projects/my-project/locations/global/apis/a/versions/v/specs/bad",
			MimeType: "application/html",
			Contents: []byte("some text"),
		},
	})

	buf := &bytes.Buffer{}
	cmd := Command()
	args := []string{"projects/my-project"}
	cmd.SetArgs(args)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", args, err)
	}

	got := new(check.CheckReport)
	if err := yaml.Unmarshal(buf.Bytes(), got); err != nil {
		t.Fatal(err)
	}

	want := &check.CheckReport{
		Id:   "check-report",
		Kind: "CheckReport",
		Problems: []*check.Problem{
			{
				Message:    `Unexpected mime_type "application/html" for contents.`,
				Suggestion: `Detected mime_type: "text/plain; charset=utf-8".`,
				Location:   `projects/my-project/locations/global/apis/a/versions/v/specs/bad::MimeType`,
				RuleId:     `registry::0111::mime-type-detected-contents`,
				Severity:   check.Problem_WARNING,
			},
		},
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(check.CheckReport), "create_time"),
	}
	if diff := cmp.Diff(want, got, opts); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestExitCode(t *testing.T) {
	ctx := context.Background()
	grpctest.SetupRegistry(ctx, t, "my-project", []seeder.RegistryResource{
		&rpc.ApiSpec{
			Name:     "projects/my-project/locations/global/apis/a/versions/v/specs/bad",
			MimeType: "application/html",
			Contents: []byte("some text"),
		},
	})

	// problem >= warning
	buf := &bytes.Buffer{}
	cmd := Command()
	args := []string{"projects/my-project", "--error-level", "WARNING"}
	cmd.SetArgs(args)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected err")
	}
	last := lastLine(buf)
	want := `Error: exceeded designated error-level "WARNING"`
	if last != want {
		t.Errorf("want: %q, got: %q", want, last)
	}

	// problem < error
	args = []string{"projects/my-project", "--error-level", "ERROR"}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
}

func lastLine(buf *bytes.Buffer) string {
	s := bufio.NewScanner(buf)
	last := ""
	for s.Scan() {
		last = s.Text()
	}
	return last
}
