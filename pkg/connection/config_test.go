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

package connection

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func cleanConfigDir(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	origConfigPath := ConfigPath
	ConfigPath = tmpDir
	return func() {
		ConfigPath = origConfigPath
		os.RemoveAll(tmpDir)
	}
}

func TestActiveSettings(t *testing.T) {
	defer cleanConfigDir(t)()
	t.Setenv("APG_REGISTRY_ADDRESS", "")

	// missing active file
	config, err := ActiveConfig()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// missing config file
	f := filepath.Join(ConfigPath, ActiveConfigPointerFilename)
	err = ioutil.WriteFile(f, []byte("missing"), os.FileMode(0644)) // rw,r,r
	if err != nil {
		t.Fatal(err)
	}
	config, err = ActiveConfig()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// invalid config file
	err = Config{}.Write("invalid")
	if err != nil {
		t.Fatal(err)
	}
	err = ActivateConfig("invalid")
	if err != nil {
		t.Fatal(err)
	}
	config, err = ActiveConfig()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// good config file
	config = Config{
		Address:  "localhost:8080",
		Insecure: true,
		Token:    "unstored",
	}
	err = config.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	err = ActivateConfig("good")
	if err != nil {
		t.Fatal(err)
	}
	got, err := ActiveConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.Token = "" // should not have been stored
	if diff := cmp.Diff(config, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsEnvVars(t *testing.T) {
	defer cleanConfigDir(t)()

	want := Config{
		Address:  "localhost:8080",
		Insecure: true,
		Token:    "mytoken",
	}
	t.Setenv("APG_REGISTRY_ADDRESS", want.Address)
	t.Setenv("APG_REGISTRY_INSECURE", strconv.FormatBool(want.Insecure))
	t.Setenv("APG_REGISTRY_TOKEN", want.Token)

	got, err := ActiveConfig()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	// good config file
	err = Config{
		Address: "overridden",
		Token:   "overridden",
	}.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	err = ActivateConfig("good")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsDirectRead(t *testing.T) {
	defer cleanConfigDir(t)()

	config := Config{
		Address:  "localhost:8080",
		Insecure: true,
	}
	err := config.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	configFile := filepath.Join(ConfigPath, "good")

	args := []string{
		"test",
		"--config", configFile,
	}
	Flags = createFlagSet()
	defer func() { Flags = createFlagSet() }()
	err = Flags.Parse(args)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ActiveConfig()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(config, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsFlags(t *testing.T) {
	defer cleanConfigDir(t)()

	want := Config{
		Address:  "localhost:8080",
		Insecure: true,
		Token:    "mytoken",
	}
	args := []string{
		"test",
		"--registry.address", want.Address,
		"--registry.insecure", strconv.FormatBool(want.Insecure),
		"--registry.token", want.Token,
	}
	Flags = createFlagSet()
	defer func() { Flags = createFlagSet() }()
	err := Flags.Parse(args)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ActiveConfig()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	// good config file
	err = Config{
		Address: "overridden",
		Token:   "overridden",
	}.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	err = ActivateConfig("good")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestMap(t *testing.T) {
	defer cleanConfigDir(t)()

	c := Config{
		Address:  "address",
		Insecure: true,
		Token:    "token",
	}
	want := map[string]interface{}{
		"address":  "address",
		"insecure": true,
		"token":    "token",
	}
	m, err := c.AsMap()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, m); diff != "" {
		t.Fatalf("unexpected diff: (-want +got):\n%s", diff)
	}

	c2 := Config{}
	err = c2.FromMap(m)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(c, c2); diff != "" {
		t.Fatalf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestAllConfigs(t *testing.T) {
	defer cleanConfigDir(t)()

	config1 := Config{
		Address:  "localhost:8080",
		Insecure: true,
	}
	err := config1.Write("config1")
	if err != nil {
		t.Fatal(err)
	}
	err = ActivateConfig("config1")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]Config{
		"config1": config1,
	}
	got, err := AllConfigs()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	config2 := Config{
		Address:  "remote:8888",
		Insecure: false,
	}
	err = config2.Write("config2")
	if err != nil {
		t.Fatal(err)
	}

	want = map[string]Config{
		"config1": config1,
		"config2": config2,
	}
	got, err = AllConfigs()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestDeleteConfig(t *testing.T) {
	defer cleanConfigDir(t)()

	config := Config{
		Address:  "localhost:8080",
		Insecure: true,
	}
	err := config.Write("config1")
	if err != nil {
		t.Fatal(err)
	}
	err = config.Write("config2")
	if err != nil {
		t.Fatal(err)
	}
	err = ActivateConfig("config1")
	if err != nil {
		t.Fatal(err)
	}

	err = DeleteConfig("config1")
	if err != CannotDeleteActiveError {
		t.Errorf("expected: %v", CannotDeleteActiveError)
	}

	err = DeleteConfig("config2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = DeleteConfig(ActiveConfigPointerFilename)
	if err != ReservedConfigNameError {
		t.Errorf("expected error: %v", ReservedConfigNameError)
	}
}

func TestWriteInvalidNames(t *testing.T) {
	defer cleanConfigDir(t)()

	err := Config{}.Write(ActiveConfigPointerFilename)
	if err != ReservedConfigNameError {
		t.Errorf("expected error: %v", ReservedConfigNameError)
	}

	err = Config{}.Write(filepath.Join("foo", "bar"))
	if err == nil {
		t.Errorf("expected error: %v", ReservedConfigNameError)
	}
}
