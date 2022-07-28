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
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

func TestCommand(t *testing.T) {
	if cmd := Command(); cmd == nil {
		t.Error("cmd not returned")
	}
}

func cleanConfigDir(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	origConfigPath := connection.ConfigPath
	connection.ConfigPath = tmpDir
	return func() {
		connection.ConfigPath = origConfigPath
	}
}

func TestNoActiveConfig(t *testing.T) {
	t.Cleanup(cleanConfigDir(t))

	checkErr := func(err error) {
		want := fmt.Errorf(`Cannot read config: No active config. Use 'registry config configurations' to manage.`)
		if err == nil {
			t.Errorf("expected error: %s", want)
		} else if diff := cmp.Diff(want.Error(), err.Error()); diff != "" {
			t.Errorf("unexpected diff: (-want +got):\n%s", diff)
		}
	}

	for _, e := range []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{"get", getCommand(), []string{"x"}},
		{"list", listCommand(), nil},
		{"set", setCommand(), []string{"x", "y"}},
		{"unset", unsetCommand(), []string{"x"}},
	} {
		t.Run(e.name, func(t *testing.T) {
			e.cmd.SetArgs(e.args)

			// missing directory
			connection.ConfigPath = filepath.Join(connection.ConfigPath, "test")
			checkErr(e.cmd.Execute())

			// empty list
			connection.ConfigPath = t.TempDir()
			checkErr(e.cmd.Execute())

			// no active
			c := connection.Config{}
			if err := c.Write("test"); err != nil {
				t.Fatal(err)
			}
			checkErr(e.cmd.Execute())
		})
	}
}

func TestConfig(t *testing.T) {
	t.Cleanup(cleanConfigDir(t))

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
	want := `Config has no property "test".`
	if err := cmd.Execute(); err.Error() != want {
		t.Errorf("expected missing property: %q", "test")
	}

	cmd = setCommand()
	cmd.SetArgs([]string{"address", "test"})
	want = "Updated property \"address\".\n"
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal()
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = setCommand()
	cmd.SetArgs([]string{"insecure", "true"})
	want = "Updated property \"insecure\".\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal()
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"test"})
	want = `Config has no property "test".`
	if err := cmd.Execute(); err.Error() != want {
		t.Errorf("expected missing property: %q", "test")
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"address"})
	want = "test\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal()
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"insecure"})
	want = "true\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal()
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = listCommand()
	want = `address = test
insecure = true

Your active configuration is: "active".
`
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal()
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = unsetCommand()
	cmd.SetArgs([]string{"address"})
	want = "Unset property \"address\".\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal()
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = unsetCommand()
	cmd.SetArgs([]string{"insecure"})
	want = "Unset property \"insecure\".\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal()
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = listCommand()
	want = `insecure = false

Your active configuration is: "active".
`
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal()
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func capture(cmd *cobra.Command, input string) (string, error) {
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader(input))

	if err := cmd.Execute(); err != nil {
		return "", err
	}

	return out.String(), nil
}
