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

package test

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"
)

// Capture sets input and output streams on a command, streams the
// input to the command (similar to a response file), and captures
// the standard output of the command as a string.
func Capture(cmd *cobra.Command, input string) (string, error) {
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader(input))

	if err := cmd.Execute(); err != nil {
		return "", err
	}

	return out.String(), nil
}
