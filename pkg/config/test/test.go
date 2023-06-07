// Copyright 2020 Google LLC.
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

package test

import (
	"testing"

	"github.com/apigee/registry/pkg/config"
)

// CleanConfigDir creates a temp config dir
// and sets the config to that directory.
// It returns a func() that should be passed
// to t.Cleanup(). Use in tests like so:
// t.Cleanup(test.CleanConfigDir(t))
func CleanConfigDir(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	origConfigPath := config.Directory
	config.Directory = tmpDir
	return func() {
		config.Directory = origConfigPath
	}
}
