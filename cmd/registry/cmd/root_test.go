// Copyright 2023 Google LLC.
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
	"flag"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestCommandMetadata(t *testing.T) {
	checkCommand(t, Command(), "")
}

var github = flag.Bool("github", false, "perform tests that check resources on GitHub")

func checkCommand(t *testing.T, cmd *cobra.Command, prefix string) {
	name := prefix + cmd.Name()
	t.Run(name, func(t *testing.T) {
		// Check short descriptions.
		short := cmd.Short
		if short == "" {
			t.Errorf("%q has no short description.", name)
		} else {
			if strings.HasSuffix(short, ".") {
				t.Errorf("%q short description must not end with a period.", name)
			}
			first := []rune(short)[0]
			if unicode.IsLower(first) {
				t.Errorf("%q short description must not begin with a lower case letter.", name)
			}
			if strings.Contains(short, "the registry") {
				t.Errorf("%q short description must refer to `the API Registry` instead of `the registry`.", name)
			}
		}
		// Perform additional checks on leaf-level commands.
		if len(cmd.Commands()) == 0 {
			args := cmd.Args
			if args == nil {
				t.Errorf("%q has an empty 'Args' field.", name)
			}
			use := cmd.Use
			if use == "" {
				t.Errorf("%q has an empty 'Use' field.", name)
			}
			if args != nil && use != "" {
				parts := strings.Split(cmd.Use, " ")
				err := args(cmd, []string{"one"})
				if err == nil {
					if len(parts) == 0 {
						t.Errorf("%q accepts arguments that are not listed in 'Use' field.", name)
					}
				}
			}

			if *github {
				// Does this command have a page on the wiki?
				client := &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				}
				wikiUrl := "https://github.com/apigee/registry/wiki/" + strings.ReplaceAll(name, " ", "-")
				res, err := client.Get(wikiUrl)
				if err != nil {
					t.Logf("error making http request: %s", err)
				} else if res.StatusCode != 200 {
					t.Logf("%s does not have a wiki page", name)
				}
			}

			// Check field usage messages.
			flags := cmd.LocalFlags()
			flags.VisitAll(func(f *pflag.Flag) {
				if f.Usage == "" {
					t.Errorf("%q %q flag usage must not be empty.", name, f.Name)
				} else {
					// Check if the first word is uppercase and not an initialism.
					if words := strings.Split(f.Usage, " "); unicode.IsUpper([]rune(f.Usage)[0]) && strings.ToUpper(words[0]) != words[0] {
						t.Errorf("%q %q flag usage must not begin with an upper case letter.", name, f.Name)
					}
					if strings.HasSuffix(f.Usage, ".") {
						t.Errorf("%q %q flag usage description must not end with a period.", name, f.Name)
					}
					re := regexp.MustCompile(`\(.*,.*\)`)
					m := re.FindStringSubmatch(f.Usage)
					if m != nil {
						t.Errorf("%q %q flag usage seems to contain an enum %q, which should be of the form (a|b|c).", name, f.Name, m[0])
					}
					// flag-specific checks
					commonFlags := []struct {
						name      string
						shorthand string
						usage     string
					}{
						{
							name:      "jobs",
							shorthand: "j",
							usage:     "number of actions to perform concurrently",
						},
						{
							name:      "force",
							shorthand: "f",
							usage:     "force deletion of child resources",
						},
						{
							name:      "filter",
							shorthand: "",
							usage:     "filter selected resources",
						},
					}
					for _, c := range commonFlags {
						if f.Name == c.name {
							if f.Shorthand != c.shorthand {
								t.Errorf("%q %q flag must have %q shorthand.", name, f.Name, c.shorthand)
							}
							if f.Usage != c.usage {
								t.Errorf("%q %q flag must have the standard usage %q.", name, f.Name, c.usage)
							}
						}
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
