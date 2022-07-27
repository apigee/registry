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
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

// Capture sets input and output streams on a command, streams the
// input to the command (similar to a response file), and captures
// the standard output of the command as a string.
func Capture(cmd *cobra.Command, input string) (string, error) {
	if input != "" {
		r, w, err := os.Pipe()
		if err != nil {
			return "", err
		}

		_, err = io.WriteString(w, input)
		if err != nil {
			return "", err
		}

		cmd.SetIn(r)
	}

	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	cmd.SetOut(w)

	err = cmd.Execute()
	if err != nil {
		return "", err
	}
	_ = w.Close()
	got, err := ioutil.ReadAll(r)
	_ = r.Close()
	if err != nil {
		return "", err
	}
	return string(got), err
}
