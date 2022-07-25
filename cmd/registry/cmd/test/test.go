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

// Capture captures stdin / stdout for a command. Input can provided
// to the cmd if there are user inputs, the output is returned.
func Capture(cmd *cobra.Command, input string) (string, error) {
	if input != "" {
		r, w, err := os.Pipe()
		if err != nil {
			return "", err
		}
		os.Stdout = w

		_, err = io.WriteString(w, input)
		if err != nil {
			return "", err
		}

		origStdin := os.Stdin
		defer func() { os.Stdin = origStdin }()
		os.Stdin = r
	}

	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

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
