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

package configurations

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apigee/registry/pkg/config"
	"github.com/apigee/registry/pkg/config/test"
	"github.com/google/go-cmp/cmp"
)

func TestCommand(t *testing.T) {
	if cmd := Command(); cmd == nil {
		t.Error("cmd not returned")
	}
}

func TestNoConfigurations(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	// missing directory
	config.Directory = filepath.Join(config.Directory, "test")
	cmd := listCommand()
	want := "You don't have any configurations. Run 'registry config configurations create' to create a configuration.\n"
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	// empty list
	config.Directory = t.TempDir()
	want = "You don't have any configurations. Run 'registry config configurations create' to create a configuration.\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestConfigurations(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))
	t.Setenv("REGISTRY_ADDRESS", "")
	t.Setenv("REGISTRY_INSECURE", "")

	cmd := createCommand()
	cmd.SetArgs([]string{"config1"})
	want := `Created "config1".
Activated "config1".
`
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd.SetArgs([]string{"config2"})
	want = `Created "config2".
Activated "config2".
`
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	name, err := config.ActiveName()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("config2", name); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = activateCommand()
	cmd.SetArgs([]string{"config1"})
	want = `Activated "config1".
`
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	config := config.Configuration{
		Registry: config.Registry{
			Address:  "foo",
			Insecure: true,
		},
	}
	err = config.Write("config2")
	if err != nil {
		t.Fatal(err)
	}

	cmd = listCommand()
	cmd.SetArgs([]string{})
	want = `NAME     IS_ACTIVE  ADDRESS  INSECURE
config1  true                false
config2  false      foo      true
`
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = describeCommand()
	cmd.SetArgs([]string{"config2"})
	want = `is_active: false
name: config2
properties:
  registry.address: foo
  registry.insecure: true
`
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = deleteCommand()
	cmd.SetArgs([]string{"config1"})
	cmd.SetIn(strings.NewReader("Y\n"))
	want = "cannot delete config \"config1\": cannot delete active configuration"
	if err = cmd.Execute(); err == nil || err.Error() != want {
		t.Errorf("unexpected error, want: %s, got: %s", want, err.Error())
	}

	cmd = deleteCommand()
	cmd.SetArgs([]string{"config2"})
	cmd.SetIn(strings.NewReader("N\n"))
	want = "aborted by user"
	if err = cmd.Execute(); err == nil || err.Error() != want {
		t.Errorf("unexpected error, want: %s, got: %s", want, err.Error())
	}

	cmd = deleteCommand()
	cmd.SetArgs([]string{"config2"})
	cmd.SetIn(strings.NewReader("Y\n"))
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	want = `The following configs will be deleted:
 - config2
Do you want to continue (Y/n)? Deleted "config2".
`
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestCreateConfiguration(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	want := config.Configuration{
		Registry: config.Registry{
			Address:  "address",
			Insecure: true,
		},
	}

	cmd := createCommand()
	cmd.PersistentFlags().AddFlagSet(config.Flags)
	// cmd.PersistentFlags().AddFlagSet(config.CreateFlagSet())
	cmd.SetArgs([]string{"config1", "--registry.address=address", "--registry.insecure=true"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	c, err := config.Read("config1")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, c); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = createCommand()
	cmd.PersistentFlags().AddFlagSet(config.Flags)
	cmd.SetArgs([]string{"config2"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if c, err = config.Read("config2"); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, c); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = createCommand()
	cmd.PersistentFlags().AddFlagSet(config.Flags)
	cmd.SetArgs([]string{"config3", "--registry.location", "location"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if c, err = config.Read("config3"); err != nil {
		t.Fatal(err)
	}
	want.Registry.Location = "location"
	if diff := cmp.Diff(want, c); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}
