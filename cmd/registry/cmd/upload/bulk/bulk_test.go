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
package bulk

import (
	"context"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/server/registry"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestParentFromDeprecatedProjectIdFlag(t *testing.T) {
	want := "projects/sample/locations/global"
	cmd := Command()
	cmd.SetContext(context.Background())
	if err := cmd.ParseFlags([]string{"--project-id", "sample"}); err != nil {
		t.Fatalf("Failed to parse flags")
	}
	parent, err := getParent(cmd)
	if err != nil {
		t.Errorf("Get parent unexpectedly failed with error: %s", err)
	}
	if parent != want {
		t.Errorf("Get parent: wanted %s, got %s", want, parent)
	}
}

func TestParentFromParentFlag(t *testing.T) {
	want := "projects/sample/locations/global"
	cmd := Command()
	cmd.SetContext(context.Background())
	if err := cmd.ParseFlags([]string{"--parent", want}); err != nil {
		t.Fatalf("Failed to parse flags")
	}
	parent, err := getParent(cmd)
	if err != nil {
		t.Errorf("Get parent unexpectedly failed with error: %s", err)
	}
	if parent != want {
		t.Errorf("Get parent: wanted %s, got %s", want, parent)
	}
}

func TestParentFromConfiguration(t *testing.T) {
	tests := []struct {
		desc       string
		projectId  string
		locationId string
		want       string
		fails      bool
	}{
		{
			desc:       "configured with specified location",
			projectId:  "sample",
			locationId: "someplace",
			want:       "projects/sample/locations/someplace",
		},
		{
			desc:       "configured with default location",
			projectId:  "sample",
			locationId: "",
			want:       "projects/sample/locations/global",
		},
		{
			desc:       "configured with no location",
			projectId:  "",
			locationId: "",
			want:       "",
			fails:      true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config, err := connection.ActiveConfig()
			if err != nil {
				t.Fatalf("Setup: Failed to get registry configuration: %s", err)
			}
			config.Project = test.projectId
			config.Location = test.locationId
			connection.SetConfig(config)
			cmd := Command()
			cmd.SetContext(context.Background())
			if err := cmd.ParseFlags([]string{}); err != nil {
				t.Fatalf("Failed to parse flags")
			}
			parent, err := getParent(cmd)
			if test.fails && err == nil {
				t.Errorf("Get parent was expected to fail and didn't")
			} else if !test.fails && err != nil {
				t.Errorf("Get parent unexpectedly failed with error: %s", err)
			} else if parent != test.want {
				t.Errorf("Get parent: wanted %s, got %s", test.want, parent)
			}
		})
	}
}
