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

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/apigee/registry/cmd/registry/cmd"
	"github.com/apigee/registry/pkg/log"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const pluginPrefix string = "registry-"

var pluginExclusions []*regexp.Regexp

func init() {
	for _, r := range []string{
		"^server$",
		"^basenames$",
		"^encode-spec$",
		"^decode-spec$",
		"^graphql$",
		"^lint-.*$",
	} {
		pluginExclusions = append(pluginExclusions, regexp.MustCompile(r))
	}
}

func main() {
	// Bind a logger instance to the local context with metadata for outbound requests.
	logger := log.NewLogger(log.DebugLevel)
	ctx := log.NewOutboundContext(log.NewContext(context.Background(), logger), log.Metadata{
		UID: fmt.Sprintf("%.8s", uuid.New()),
	})

	cmd := cmd.Command()
	cmd.SetUsageTemplate(usageTemplate)
	cobra.AddTemplateFunc("Plugins", plugins)

	staySilent := cmd.SilenceErrors
	cmd.SilenceErrors = true // don't print "unknown command" error
	if err := cmd.ExecuteContext(ctx); err != nil {
		if strings.HasPrefix(err.Error(), "unknown command") && contains(plugins(), os.Args[1]) {
			exCmd, err := exec.LookPath(pluginPrefix + os.Args[1])
			if err == nil {
				if err := syscall.Exec(exCmd, append([]string{exCmd}, os.Args[2:]...), os.Environ()); err != nil {
					fmt.Fprintf(os.Stderr, "Command finished with error: %v", err)
					os.Exit(127)
				}
			}
		}
		if !staySilent {
			cmd.PrintErrln("Error:", err.Error())
			cmd.PrintErrf("Run '%v --help' for usage.\n", cmd.CommandPath())
		}
		os.Exit(1)
	}
}

// adds `Available Plugins` and `Need more help?` sections to default
const usageTemplate = `Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if gt (len Plugins) 0}}

Available Plugins:
{{- range Plugins}}
  {{.}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}

Need more help?
https://github.com/apigee/registry/wiki
`

// list of executable files on env.PATH matching pluginPrefix,
// excluding pluginExclusions
func plugins() []string {
	plugs := []string{}
	paths := filepath.SplitList(os.Getenv("PATH"))
	for _, path := range paths {
		if files, err := os.ReadDir(path); err == nil {
		skip:
			for _, file := range files {
				filename := file.Name()
				if strings.HasPrefix(filename, pluginPrefix) {
					plug := strings.TrimPrefix(filename, pluginPrefix)
					for _, re := range pluginExclusions {
						if re.MatchString(plug) {
							continue skip
						}
					}
					if !contains(plugs, plug) {
						if info, err := os.Stat(filepath.Join(path, filename)); err == nil {
							if info.Mode()&0o111 != 0 {
								plugs = append(plugs, plug)
							}
						}
					}
				}
			}
		}
	}
	return plugs
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
