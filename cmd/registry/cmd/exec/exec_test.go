// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"bytes"
	"context"
	"testing"
)

// COMMAND should be passed as a single arg.
// Example: registry exec "echo test"
func TestExec(t *testing.T) {
	tests := []struct {
		desc    string
		command string
		want    string
	}{
		{
			desc:    "single cmd arg",
			command: "echo test",
			want:    "test\n",
		},
		{
			desc:    "multiple cmd arg",
			command: "echo sample test",
			want:    "sample test\n",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			cmd := Command(ctx)
			out := bytes.NewBuffer(make([]byte, 0))
			args := []string{test.command}
			cmd.SetArgs(args)
			cmd.SetOutput(out)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}

			output := out.String()
			// The exec command should execute all the args passed to it
			// Make sure that the output produced is "sample test" and not only "sample" for multi-arg case
			if output != test.want {
				t.Fatalf("Execute() with args %v generated unexpected output, want: %q got: %q", args, test.want, output)
			}
		})
	}

}
