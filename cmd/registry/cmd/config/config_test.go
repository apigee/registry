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

package config

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/pkg/config"
	"github.com/apigee/registry/pkg/config/test"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

func TestCommand(t *testing.T) {
	if cmd := Command(); cmd == nil {
		t.Error("cmd not returned")
	}
}

func TestNoActiveConfig(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	checkErr := func(t *testing.T, err error) {
		t.Helper()
		want := fmt.Errorf(`no active configuration, use 'registry config configurations' to manage`)
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
			config.Directory = filepath.Join(config.Directory, "test")
			checkErr(t, e.cmd.Execute())

			// empty list
			config.Directory = t.TempDir()
			checkErr(t, e.cmd.Execute())

			// no active
			c := config.Configuration{}
			if err := c.Write("test"); err != nil {
				t.Fatal(err)
			}
			checkErr(t, e.cmd.Execute())
		})
	}
}

func TestConfig(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	c := config.Configuration{}
	err := c.Write("active")
	if err != nil {
		t.Fatal(err)
	}
	err = config.Activate("active")
	if err != nil {
		t.Fatal(err)
	}

	cmd := setCommand()
	cmd.SetArgs([]string{"test", "test"})
	if err := cmd.Execute(); !errors.Is(err, config.UnknownPropertyError{Property: "test"}) {
		t.Errorf("expected UnknownPropertyError")
	}

	cmd = setCommand()
	cmd.SetArgs([]string{"registry.address", "test"})
	want := "Updated property \"registry.address\".\n"
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = setCommand()
	cmd.SetArgs([]string{"registry.insecure", "true"})
	want = "Updated property \"registry.insecure\".\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"test"})
	if err := cmd.Execute(); !errors.Is(err, config.UnknownPropertyError{Property: "test"}) {
		t.Errorf("expected UnknownPropertyError")
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"registry.address"})
	want = "test\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = getCommand()
	cmd.SetArgs([]string{"registry.insecure"})
	want = "true\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = listCommand()
	cmd.SetArgs([]string{})
	want = `registry.address = test
registry.insecure = true

Your active configuration is: "active".
`
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = unsetCommand()
	cmd.SetArgs([]string{"registry.address"})
	want = "Unset property \"registry.address\".\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = unsetCommand()
	cmd.SetArgs([]string{"registry.insecure"})
	want = "Unset property \"registry.insecure\".\n"
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	cmd = listCommand()
	cmd.SetArgs([]string{})
	want = `registry.insecure = false

Your active configuration is: "active".
`
	out = new(bytes.Buffer)
	cmd.SetOut(out)
	if err = cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}
