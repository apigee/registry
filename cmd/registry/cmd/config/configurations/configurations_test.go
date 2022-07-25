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

package configurations

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

func TestConfigurations(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigPath := connection.ConfigPath
	connection.ConfigPath = tmpDir
	defer func() { connection.ConfigPath = origConfigPath }()

	cmd := createCommand()
	cmd.SetArgs([]string{"config1"})
	got, err := test.Capture(cmd, "")
	if err != nil {
		t.Fatal(err)
	}
	want := `Created "config1".
Activated "config1".
`
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd.SetArgs([]string{"config2"})
	got, err = test.Capture(cmd, "")
	if err != nil {
		t.Fatal(err)
	}
	want = `Created "config2".
Activated "config2".
`
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	name, err := connection.ActiveConfigName()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("config2", name); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = activateCommand()
	cmd.SetArgs([]string{"config1"})
	got, err = test.Capture(cmd, "")
	if err != nil {
		t.Fatal(err)
	}
	want = `Activated "config1".
`
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	config := connection.Config{
		Address:  "foo",
		Insecure: true,
	}
	err = config.Write("config2")
	if err != nil {
		t.Fatal(err)
	}

	cmd = listCommand()
	cmd.SetArgs([]string{})
	got, err = test.Capture(cmd, "")
	if err != nil {
		t.Fatal(err)
	}
	want = `NAME     IS_ACTIVE  ADDRESS  INSECURE
config1  true                false
config2  false      foo      true
`
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = describeCommand()
	cmd.SetArgs([]string{"config2"})
	got, err = test.Capture(cmd, "")
	if err != nil {
		t.Fatal(err)
	}
	want = `is_active: false
name: config2
properties:
  address: foo
  insecure: true
`
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = deleteCommand()
	cmd.SetArgs([]string{"config1"})
	input := "Y\n"
	_, err = test.Capture(cmd, input)
	want = "Cannot delete config \"config1\": Cannot delete active configuration."
	if err == nil && err.Error() != want {
		t.Errorf("expected error: %s", want)
	}

	cmd = deleteCommand()
	cmd.SetArgs([]string{"config2"})
	input = "N\n"
	_, err = test.Capture(cmd, input)
	want = "Aborted by user."
	if err != nil && err.Error() != want {
		t.Errorf("expected error: %s", want)
	}

	cmd = deleteCommand()
	cmd.SetArgs([]string{"config2"})
	input = "Y\n"
	got, err = test.Capture(cmd, input)
	if err != nil {
		t.Fatal(err)
	}
	want = `The following configs will be deleted:
 - config2
Do you want to continue (Y/n)? Deleted "config2".
`
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}
