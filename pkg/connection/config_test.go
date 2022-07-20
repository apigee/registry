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
	"github.com/spf13/pflag"
)

func TestMain(m *testing.M) {
	tmpDir, err := ioutil.TempDir("", "registry")
	if err != nil {
		panic("can't create temp dir")
	}
	configPath = tmpDir
	code := m.Run()
	os.Exit(code)
}

func TestActiveSettings(t *testing.T) {
	// ensure outside env var doesn't affect this test
	addrEnv := os.Getenv("APG_REGISTRY_ADDRESS")
	if err := os.Unsetenv("APG_REGISTRY_ADDRESS"); err != nil {
		t.Fatal(err)
	}
	defer os.Setenv("APG_REGISTRY_ADDRESS", addrEnv)

	// missing active file
	config, err := ActiveConfig()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// missing config file
	f := filepath.Join(configPath, activeConfigPointerFilename)
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
		t.Errorf("unexpected error: %s", err)
	}
	config.Token = "" // should not have been stored
	if diff := cmp.Diff(config, got); diff != "" {
		t.Errorf("activeSettings returned unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsEnvVars(t *testing.T) {
	// no config files
	want := Config{
		Address:  "localhost:8080",
		Insecure: true,
		Token:    "mytoken",
	}
	os.Setenv("APG_REGISTRY_ADDRESS", want.Address)
	defer os.Unsetenv("APG_REGISTRY_ADDRESS")
	os.Setenv("APG_REGISTRY_INSECURE", strconv.FormatBool(want.Insecure))
	defer os.Unsetenv("APG_REGISTRY_INSECURE")
	os.Setenv("APG_REGISTRY_TOKEN", want.Token)
	defer os.Unsetenv("APG_REGISTRY_TOKEN")

	got, err := ActiveConfig()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("activeSettings returned unexpected diff: (-want +got):\n%s", diff)
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
		t.Errorf("activeSettings returned unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsDirectRead(t *testing.T) {
	config := Config{
		Address:  "localhost:8080",
		Insecure: true,
	}
	err := config.Write("good")
	if err != nil {
		t.Fatal(err)
	}
	configFile := filepath.Join(configPath, "good")

	args := os.Args
	defer func() { os.Args = args }()
	os.Args = []string{
		"test",
		"--config", configFile,
	}
	pflag.CommandLine.AddFlagSet(Flags)
	pflag.Parse()
	got, err := ActiveConfig()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(config, got); diff != "" {
		t.Errorf("activeSettings returned unexpected diff: (-want +got):\n%s", diff)
	}
}

func TestSettingsFlags(t *testing.T) {
	want := Config{
		Address:  "localhost:8080",
		Insecure: true,
		Token:    "mytoken",
	}
	args := os.Args
	defer func() { os.Args = args }()
	os.Args = []string{
		"test",
		"--registry.address", want.Address,
		"--registry.insecure", strconv.FormatBool(want.Insecure),
		"--registry.token", want.Token,
	}
	pflag.CommandLine.AddFlagSet(Flags)
	pflag.Parse()
	got, err := ActiveConfig()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("activeSettings returned unexpected diff: (-want +got):\n%s", diff)
	}
}
