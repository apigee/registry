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

package config

import (
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/test"
	"github.com/apigee/registry/pkg/connection"
	"github.com/google/go-cmp/cmp"
)

func TestCommand(t *testing.T) {
	if cmd := Command(); cmd == nil {
		t.Error("cmd not returned")
	}
}

func TestConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigPath := connection.ConfigPath
	connection.ConfigPath = tmpDir
	defer func() { connection.ConfigPath = origConfigPath }()

	s := connection.Config{}
	err := s.Write("active")
	if err != nil {
		t.Fatal(err)
	}
	err = connection.ActivateConfig("active")
	if err != nil {
		t.Fatal(err)
	}

	cmd := setCommand()
	cmd.SetArgs([]string{"test", "test"})
	_, err = test.Capture(cmd, "")
	if err.Error() != `Config has no property "test".` {
		t.Errorf("expected missing property: %q", "test")
	}

	cmd = setCommand()
	cmd.SetArgs([]string{"address", "test"})
	got, err := test.Capture(cmd, "")
	want := "Updated property \"address\".\n"
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = setCommand()
	cmd.SetArgs([]string{"insecure", "true"})
	got, err = test.Capture(cmd, "")
	want = "Updated property \"insecure\".\n"
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"test"})
	_, err = test.Capture(cmd, "")
	if err.Error() != `Config has no property "test".` {
		t.Errorf("expected missing property: %q", "test")
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"address"})
	got, err = test.Capture(cmd, "")
	want = "test\n"
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"insecure"})
	got, err = test.Capture(cmd, "")
	want = "true\n"
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = listCommand()
	got, err = test.Capture(cmd, "")
	want = `address = test
insecure = true

Your active configuration is: "active".
`
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = unsetCommand()
	cmd.SetArgs([]string{"address"})
	got, err = test.Capture(cmd, "")
	want = "Unset property \"address\".\n"
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = unsetCommand()
	cmd.SetArgs([]string{"insecure"})
	got, err = test.Capture(cmd, "")
	want = "Unset property \"insecure\".\n"
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = listCommand()
	got, err = test.Capture(cmd, "")
	want = `insecure = false

Your active configuration is: "active".
`
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

}
