// Copyright 2023 Google LLC. All Rights Reserved.
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

package cmd

import (
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestCommandMetadata(t *testing.T) {
	checkCommand(t, Command(), "")
}

func checkCommand(t *testing.T, cmd *cobra.Command, prefix string) {
	name := prefix + cmd.Name()
	t.Run(name, func(t *testing.T) {
		// Check short descriptions.
		short := cmd.Short
		if short == "" {
			t.Logf("%q has no short description.", name)
		} else {
			if strings.HasSuffix(short, ".") {
				t.Logf("%q short description must not end with a period.", name)
			}
			first := []rune(short)[0]
			if unicode.IsLower(first) {
				t.Logf("%q short description must not begin with a lower case letter.", name)
			}
			if strings.Contains(short, "the registry") {
				t.Logf("%q short description must refer to `the API Registry` instead of `the registry`.", name)
			}
		}
		// Perform additional checks on leaf-level commands.
		if len(cmd.Commands()) == 0 {
			args := cmd.Args
			if args == nil {
				t.Logf("%q has an empty 'Args' field.", name)
			}
			use := cmd.Use
			if use == "" {
				t.Logf("%q has an empty 'Use' field.", name)
			}
			if args != nil && use != "" {
				parts := strings.Split(cmd.Use, " ")
				err := args(cmd, []string{"one"})
				if err == nil {
					if len(parts) == 0 {
						t.Logf("%q accepts arguments that are not listed in 'Use' field.", name)
					}
				}
			}
			// Check field descriptions.
			flags := cmd.Flags()
			flags.VisitAll(func(f *pflag.Flag) {
				description := f.Shorthand
				if description == "" {
					t.Logf("%q %q flag description must not be empty.", name, f.Name)
				} else {
					first := []rune(description)[0]
					if unicode.IsUpper(first) {
						t.Logf("%q %q flag description must not begin with an upper case letter.", name, f.Name)
					}
				}
			})
		}
	})
	// Check subcommands, but skip auto-generated ones. We can't easily fix those.
	if cmd.Name() == "rpc" {
		return
	}
	for _, c := range cmd.Commands() {
		checkCommand(t, c, name+" ")
	}
}
