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

package config_test

import (
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/apigee/registry/pkg/config"
	"github.com/apigee/registry/pkg/config/test"
	"github.com/google/go-cmp/cmp"
)

func TestMissingDirectory(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	config.Directory = path.Join(config.Directory, "test")
	c := config.Configuration{}
	if err := c.Write("foo"); err != nil {
		t.Fatalf("unexpected error on write: %s", err)
	}
}

func TestActiveSettings(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))
	t.Setenv("REGISTRY_ADDRESS", "")

	// missing active file
	_, err := config.Active()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// missing config file
	f := filepath.Join(config.Directory, config.ActivePointerFilename)
	err = os.WriteFile(f, []byte("missing"), os.FileMode(0644)) // rw,r,r
	if err != nil {
		t.Fatal(err)
	}
	_, err = config.Active()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// invalid config file
	err = config.Configuration{}.Write("invalid")
	if err != nil {
		t.Fatal(err)
	}
	err = config.Activate("invalid")
	if err != nil {
		t.Fatal(err)
	}
	_, err = config.Active()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// good config file
	c := config.Configuration{
		Registry: config.Registry{
			Address:  "localhost:8080",
			Insecure: true,
		},
	}
	err = c.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	err = config.Activate("good")
	if err != nil {
		t.Fatal(err)
	}
	got, err := config.Active()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(c, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	t.Setenv("REGISTRY_ADDRESS", "ignore")
	name, raw, err := config.ActiveRaw()
	if err != nil {
		t.Fatal(err)
	}
	if name != "good" {
		t.Errorf("want: %s, got: %s", "good", name)
	}
	if diff := cmp.Diff(raw, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsEnvVars(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	want := config.Configuration{
		Registry: config.Registry{
			Address:  "localhost:8080",
			Insecure: true,
			Token:    "token",
		},
	}
	t.Setenv("REGISTRY_ADDRESS", want.Registry.Address)
	t.Setenv("REGISTRY_INSECURE", strconv.FormatBool(want.Registry.Insecure))
	t.Setenv("REGISTRY_TOKEN", want.Registry.Token)

	got, err := config.Active()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	// good config file
	err = config.Configuration{
		Registry: config.Registry{
			Address: "overridden",
		},
	}.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	err = config.Activate("good")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsDirectRead(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	c := config.Configuration{
		Registry: config.Registry{
			Address:  "localhost:8080",
			Insecure: true,
			Location: "location",
			Project:  "project",
		},
	}
	err := c.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	configFile := filepath.Join(config.Directory, "good")

	args := []string{
		"test",
		"--config", configFile,
	}
	config.Flags = config.CreateFlagSet()
	defer func() { config.Flags = config.CreateFlagSet() }()
	err = config.Flags.Parse(args)
	if err != nil {
		t.Fatal(err)
	}

	got, err := config.Active()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(c, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsFlags(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	want := config.Configuration{
		Registry: config.Registry{
			Address:  "localhost:8080",
			Insecure: true,
			Location: "location",
			Project:  "project",
			Token:    "token",
		},
	}
	args := []string{
		"test",
		"--registry.address", want.Registry.Address,
		"--registry.insecure", strconv.FormatBool(want.Registry.Insecure),
		"--registry.location", want.Registry.Location,
		"--registry.project", want.Registry.Project,
		"--registry.token", want.Registry.Token,
	}
	config.Flags = config.CreateFlagSet()
	defer func() { config.Flags = config.CreateFlagSet() }()
	err := config.Flags.Parse(args)
	if err != nil {
		t.Fatal(err)
	}

	got, err := config.Active()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	// good config file
	err = config.Configuration{
		Registry: config.Registry{
			Address: "overridden",
		},
	}.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	err = config.Activate("good")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestManipulations(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	c := config.Configuration{
		Registry: config.Registry{
			Address:  "address",
			Insecure: true,
			Location: "location",
			Project:  "project",
			Token:    "token",
		},
	}
	want := map[string]interface{}{
		"registry.address":  c.Registry.Address,
		"registry.insecure": c.Registry.Insecure,
		"registry.location": c.Registry.Location,
		"registry.project":  c.Registry.Project,
		"registry.token":    c.Registry.Token,
		"token-source":      "",
	}
	m, err := c.FlatMap()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, m); diff != "" {
		t.Fatalf("unexpected diff: (-want +got):\n%s", diff)
	}

	c2 := config.Configuration{}
	if err = c2.FromMap(m); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(c, c2); diff != "" {
		t.Fatalf("unexpected diff: (-want +got):\n%s", diff)
	}

	if err = c.Set("registry.address", "new address"); err != nil {
		t.Fatal(err)
	}
	if m, err = c.FlatMap(); err != nil {
		t.Fatal(err)
	}
	want["registry.address"] = "new address"
	if diff := cmp.Diff(want, m); diff != "" {
		t.Fatalf("unexpected diff: (-want +got):\n%s", diff)
	}

	if err = c.Set("location", "new location"); err != nil {
		t.Fatal(err)
	}
	if m, err = c.FlatMap(); err != nil {
		t.Fatal(err)
	}
	want["registry.location"] = "new location"
	if diff := cmp.Diff(want, m); diff != "" {
		t.Fatalf("unexpected diff: (-want +got):\n%s", diff)
	}

	l, err := c.Get("location")
	if err != nil {
		t.Fatal(err)
	}
	if l != "new location" {
		t.Errorf("want: %s, got: %s", "new location", l)
	}

	if err = c.Unset("registry.address"); err != nil {
		t.Fatal(err)
	}
	if err = c.Unset("insecure"); err != nil {
		t.Fatal(err)
	}
	if m, err = c.FlatMap(); err != nil {
		t.Fatal(err)
	}
	want["registry.address"] = ""
	want["registry.insecure"] = false
	if diff := cmp.Diff(want, m); diff != "" {
		t.Fatalf("unexpected diff: (-want +got):\n%s", diff)
	}

	if err = c.Set("token-source", "source"); err != nil {
		t.Fatal(err)
	}
	if m, err = c.FlatMap(); err != nil {
		t.Fatal(err)
	}
	want["token-source"] = "source"
	if diff := cmp.Diff(want, m); diff != "" {
		t.Fatalf("unexpected diff: (-want +got):\n%s", diff)
	}

	ts, err := c.Get("token-source")
	if err != nil {
		t.Fatal(err)
	}
	if ts != "source" {
		t.Errorf("want: %s, got: %s", "source", l)
	}
}

func TestAllConfigs(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))
	t.Setenv("REGISTRY_ADDRESS", "")
	t.Setenv("REGISTRY_INSECURE", "")

	config1 := config.Configuration{
		Registry: config.Registry{
			Address:  "localhost:8080",
			Insecure: true,
		},
	}
	err := config1.Write("config1")
	if err != nil {
		t.Fatal(err)
	}
	err = config.Activate("config1")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]config.Configuration{
		"config1": config1,
	}
	got, err := config.Configurations()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	config2 := config.Configuration{
		Registry: config.Registry{
			Address:  "remote:8888",
			Insecure: false,
		},
	}
	err = config2.Write("config2")
	if err != nil {
		t.Fatal(err)
	}

	want = map[string]config.Configuration{
		"config1": config1,
		"config2": config2,
	}
	got, err = config.Configurations()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestDeleteConfig(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	c := config.Configuration{
		Registry: config.Registry{
			Address:  "localhost:8080",
			Insecure: true,
		},
	}
	err := c.Write("config1")
	if err != nil {
		t.Fatal(err)
	}
	err = c.Write("config2")
	if err != nil {
		t.Fatal(err)
	}
	err = config.Activate("config1")
	if err != nil {
		t.Fatal(err)
	}

	err = config.Delete("config1")
	if err != config.ErrCannotDeleteActive {
		t.Errorf("expected: %v", config.ErrCannotDeleteActive)
	}

	err = config.Delete("config2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = config.Delete(config.ActivePointerFilename)
	if err != config.ErrReservedConfigName {
		t.Errorf("expected error: %v", config.ErrReservedConfigName)
	}
}

func TestWriteInvalidNames(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))

	err := config.Configuration{}.Write(config.ActivePointerFilename)
	if err != config.ErrReservedConfigName {
		t.Errorf("expected error: %v", config.ErrReservedConfigName)
	}

	err = config.Configuration{}.Write(filepath.Join("foo", "bar"))
	if err == nil {
		t.Errorf("expected error: %v", config.ErrReservedConfigName)
	}
}

func TestProperies(t *testing.T) {
	c := config.Configuration{}

	got := c.Properties()
	want := []string{
		"registry.address",
		"registry.insecure",
		"registry.location",
		"registry.project",
		"registry.token",
		"token-source",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestResolve(t *testing.T) {
	c := config.Configuration{
		TokenSource: "echo hello",
	}

	err := c.Resolve()
	if err != nil {
		t.Fatal(err)
	}

	if c.Registry.Token != "hello" {
		t.Errorf("want: %s, got: %s", "hello", c.Registry.Token)
	}
}
