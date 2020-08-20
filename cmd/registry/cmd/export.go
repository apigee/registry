// Copyright 2020 Google LLC. All Rights Reserved.
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
	"fmt"
	"log"
	"os"
	"path"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(exportCmd)
}

func check(err error) {
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a subtree of the Registry.",
	Long:  `Export a subtree of the Registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		var name string
		if len(args) > 0 {
			name = args[0]
		}

		if m := names.ProjectRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err := getProject(ctx, client, m[0], func(message *rpc.Project) {
				docMapContent := nodeSlice()
				apisMapContent := nodeSlice()
				err = listAPIs(ctx, client, m[0], func(message *rpc.Api) {
					m := names.ApiRegexp().FindAllStringSubmatch(message.Name, -1)
					apiMapContent := nodeSlice()
					versionsMapContent := nodeSlice()
					err = listVersions(ctx, client, m[0], func(message *rpc.Version) {
						m := names.VersionRegexp().FindAllStringSubmatch(message.Name, -1)
						versionMapContent := nodeSlice()
						specsMapContent := nodeSlice()
						err = listSpecs(ctx, client, m[0], func(message *rpc.Spec) {
							specMapContent := nodeSlice()
							specMapContent = appendPair(specMapContent, "style", nodeForString(message.Style))
							specMapContent = appendPair(specMapContent, "hash", nodeForString(message.Hash))
							specMapContent = appendPair(specMapContent, "size", nodeForInt64(int64(message.SizeBytes)))
							specMapContent = appendPair(specMapContent, "revisionId", nodeForString(message.RevisionId))
							specsMapContent = appendPair(specsMapContent, path.Base(message.Name), nodeForMapping(specMapContent))
						})
						check(err)
						versionMapContent = appendPair(versionMapContent, "specs", nodeForMapping(specsMapContent))
						versionsMapContent = appendPair(versionsMapContent, path.Base(message.Name), nodeForMapping(versionMapContent))
					})
					check(err)
					apiMapContent = appendPair(apiMapContent, "versions", nodeForMapping(versionsMapContent))
					apisMapContent = appendPair(apisMapContent, path.Base(message.Name), nodeForMapping(apiMapContent))
				})
				check(err)
				docMapContent = appendPair(docMapContent, "apis", nodeForMapping(apisMapContent))
				// add list of labels
				// add list of properties
				// create the top-level document
				doc := &yaml.Node{
					Kind: yaml.DocumentNode,
					Content: []*yaml.Node{
						nodeForMapping(docMapContent),
					},
				}
				// write the doc as yaml
				b, err := yaml.Marshal(doc)
				if err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(string(b))
			})
			check(err)
		} else if m := names.ApiRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err = getAPI(ctx, client, m[0], exportAPI)
		} else if m := names.VersionRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err = getVersion(ctx, client, m[0], exportVersion)
		} else if m := names.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			_, err = getSpec(ctx, client, m[0], false, exportSpec)
		} else {
			log.Fatalf("Unsupported entity %+s", name)
		}
	},
}

func exportAPI(message *rpc.Api) {
	printMessage(message)
}

func exportVersion(message *rpc.Version) {
	printMessage(message)
}

func exportSpec(message *rpc.Spec) {
	if getContents {
		os.Stdout.Write(message.GetContents())
	} else {
		printMessage(message)
	}
}

func nodeForMapping(content []*yaml.Node) *yaml.Node {
	if content == nil {
		content = make([]*yaml.Node, 0)
	}
	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: content,
	}
}

func nodeForSequence(content []*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: content,
	}
}

func nodeForString(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
}

func nodeForBoolean(value bool) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!bool",
		Value: fmt.Sprintf("%t", value),
	}
}

func nodeForInt64(value int64) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!int",
		Value: fmt.Sprintf("%d", value),
	}
}

func nodeForFloat64(value float64) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!float",
		Value: fmt.Sprintf("%f", value),
	}
}

func appendPair(nodes []*yaml.Node, name string, value *yaml.Node) []*yaml.Node {
	nodes = append(nodes, nodeForString(name))
	nodes = append(nodes, value)
	return nodes
}

func nodeSlice() []*yaml.Node {
	return make([]*yaml.Node, 0)
}
