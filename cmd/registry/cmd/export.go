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
	"path"
	"time"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
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
	Short: "Export a subtree of the registry.",
	Long:  `Export a subtree of the registry.`,
	Args:  cobra.MinimumNArgs(1),
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

		if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
			_, err := core.GetProject(ctx, client, m, func(message *rpc.Project) {
				printDocAsYaml(docForMapping(exportProject(ctx, client, message)))
			})
			check(err)
		} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
			_, err = core.GetAPI(ctx, client, m, func(message *rpc.Api) {
				printDocAsYaml(docForMapping(exportAPI(ctx, client, message)))
			})
			check(err)
		} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
			_, err = core.GetVersion(ctx, client, m, func(message *rpc.Version) {
				printDocAsYaml(docForMapping(exportVersion(ctx, client, message)))
			})
			check(err)
		} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
			_, err = core.GetSpec(ctx, client, m, false, func(message *rpc.Spec) {
				printDocAsYaml(docForMapping(exportSpec(ctx, client, message)))
			})
			check(err)
		} else {
			log.Fatalf("Unsupported entity %+s", name)
		}
	},
}

func exportProject(ctx context.Context, client *gapic.RegistryClient, message *rpc.Project) []*yaml.Node {
	m := names.ProjectRegexp().FindStringSubmatch(message.Name)
	projectMapContent := nodeSlice()
	apisMapContent := nodeSlice()
	err := core.ListAPIs(ctx, client, m, "", func(message *rpc.Api) {
		apiMapContent := exportAPI(ctx, client, message)
		apisMapContent = appendPair(apisMapContent, path.Base(message.Name), nodeForMapping(apiMapContent))
	})
	check(err)
	projectMapContent = appendPair(projectMapContent, "apis", nodeForMapping(apisMapContent))
	return projectMapContent
}

func exportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) []*yaml.Node {
	m := names.ApiRegexp().FindStringSubmatch(message.Name)
	apiMapContent := nodeSlice()
	apiMapContent = appendPair(apiMapContent, "createTime", nodeForTime(message.CreateTime.AsTime()))
	apiMapContent = appendPair(apiMapContent, "availability", nodeForString(message.Availability))
	apiMapContent = appendPair(apiMapContent, "recommended_version", nodeForString(message.RecommendedVersion))
	versionsMapContent := nodeSlice()
	err := core.ListVersions(ctx, client, m, "", func(message *rpc.Version) {
		versionMapContent := exportVersion(ctx, client, message)
		versionsMapContent = appendPair(versionsMapContent, path.Base(message.Name), nodeForMapping(versionMapContent))
	})
	check(err)
	apiMapContent = appendPair(apiMapContent, "versions", nodeForMapping(versionsMapContent))
	labelsArrayContent := nodeSlice()
	err = core.ListLabelsForParent(ctx, client, m, func(message *rpc.Label) {
		labelsArrayContent = append(labelsArrayContent, nodeForString(path.Base(message.Name)))
	})
	check(err)
	if len(labelsArrayContent) > 0 {
		apiMapContent = appendPair(apiMapContent, "labels", nodeForSequence(labelsArrayContent))
	}
	propertiesMapContent := nodeSlice()
	err = core.ListPropertiesForParent(ctx, client, m, func(message *rpc.Property) {
		propertiesMapContent = appendPair(propertiesMapContent,
			path.Base(message.Name),
			nodeForMapping(exportProperty(ctx, client, message)))
	})
	if len(propertiesMapContent) > 0 {
		apiMapContent = appendPair(apiMapContent, "properties", nodeForMapping(propertiesMapContent))
	}
	return apiMapContent
}

func exportVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.Version) []*yaml.Node {
	m := names.VersionRegexp().FindStringSubmatch(message.Name)
	versionMapContent := nodeSlice()
	versionMapContent = appendPair(versionMapContent, "createTime", nodeForTime(message.CreateTime.AsTime()))
	versionMapContent = appendPair(versionMapContent, "state", nodeForString(message.State))
	specsMapContent := nodeSlice()
	err := core.ListSpecs(ctx, client, m, "", func(message *rpc.Spec) {
		specMapContent := exportSpec(ctx, client, message)
		specsMapContent = appendPair(specsMapContent, path.Base(message.Name), nodeForMapping(specMapContent))
		m := names.SpecRegexp().FindStringSubmatch(message.Name)
		err := core.ListSpecRevisions(ctx, client, m, "", func(message *rpc.Spec) {
			specMapContent := exportSpec(ctx, client, message)
			specsMapContent = appendPair(specsMapContent, path.Base(message.Name), nodeForMapping(specMapContent))
		})
		check(err)
	})
	check(err)
	versionMapContent = appendPair(versionMapContent, "specs", nodeForMapping(specsMapContent))
	labelsArrayContent := nodeSlice()
	err = core.ListLabelsForParent(ctx, client, m, func(message *rpc.Label) {
		labelsArrayContent = append(labelsArrayContent, nodeForString(path.Base(message.Name)))
	})
	if len(labelsArrayContent) > 0 {
		versionMapContent = appendPair(versionMapContent, "labels", nodeForSequence(labelsArrayContent))
	}
	propertiesMapContent := nodeSlice()
	err = core.ListPropertiesForParent(ctx, client, m, func(message *rpc.Property) {
		propertiesMapContent = appendPair(propertiesMapContent,
			path.Base(message.Name),
			nodeForMapping(exportProperty(ctx, client, message)))
	})
	if len(propertiesMapContent) > 0 {
		versionMapContent = appendPair(versionMapContent, "properties", nodeForMapping(propertiesMapContent))
	}
	return versionMapContent
}

func exportSpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.Spec) []*yaml.Node {
	specMapContent := nodeSlice()
	specMapContent = appendPair(specMapContent, "style", nodeForString(message.Style))
	specMapContent = appendPair(specMapContent, "hash", nodeForString(message.Hash))
	specMapContent = appendPair(specMapContent, "size", nodeForInt64(int64(message.SizeBytes)))
	specMapContent = appendPair(specMapContent, "createTime", nodeForTime(message.CreateTime.AsTime()))
	specMapContent = appendPair(specMapContent, "revisionId", nodeForString(message.RevisionId))
	return specMapContent
}

func exportProperty(ctx context.Context, client *gapic.RegistryClient, message *rpc.Property) []*yaml.Node {
	propertyMapContent := nodeSlice()
	switch v := message.Value.(type) {
	case *rpc.Property_StringValue:
		propertyMapContent = appendPair(propertyMapContent, "value", nodeForString(
			v.StringValue))
	case *rpc.Property_Int64Value:
		propertyMapContent = appendPair(propertyMapContent, "value", nodeForInt64(
			v.Int64Value))
	default:
		propertyMapContent = appendPair(propertyMapContent, "value", nodeForString(
			fmt.Sprintf("%+v", message.Value)))
	}
	propertyMapContent = appendPair(propertyMapContent, "createTime", nodeForTime(message.CreateTime.AsTime()))
	return propertyMapContent
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

func nodeForTime(t time.Time) *yaml.Node {
	s, _ := t.MarshalText()
	return nodeForString(string(s))
}

func appendPair(nodes []*yaml.Node, name string, value *yaml.Node) []*yaml.Node {
	nodes = append(nodes, nodeForString(name))
	nodes = append(nodes, value)
	return nodes
}

func nodeSlice() []*yaml.Node {
	return make([]*yaml.Node, 0)
}

func docForMapping(nodes []*yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{
			nodeForMapping(nodes),
		},
	}
}

func printDocAsYaml(doc *yaml.Node) {
	b, err := yaml.Marshal(doc)
	check(err)
	fmt.Println(string(b))
}
