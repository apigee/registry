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

package yaml

import (
	"context"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"gopkg.in/yaml.v3"
)

func check(err error) {
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}

// exportYAMLForProject writes a project as a YAML file.
func exportYAMLForProject(ctx context.Context, client *gapic.RegistryClient, message *rpc.Project) {
	printDocAsYaml(docForMapping(exportProject(ctx, client, message)))
}

// exportYAMLForAPI writes a project as a YAML file.
func exportYAMLForAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) {
	printDocAsYaml(docForMapping(exportAPI(ctx, client, message)))
}

// exportYAMLForVersion writes a project as a YAML file.
func exportYAMLForVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) {
	printDocAsYaml(docForMapping(exportVersion(ctx, client, message)))
}

// exportYAMLForSpec writes a project as a YAML file.
func exportYAMLForSpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec) {
	printDocAsYaml(docForMapping(exportSpec(ctx, client, message)))
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
	err := core.ListVersions(ctx, client, m, "", func(message *rpc.ApiVersion) {
		versionMapContent := exportVersion(ctx, client, message)
		versionsMapContent = appendPair(versionsMapContent, path.Base(message.Name), nodeForMapping(versionMapContent))
	})
	check(err)
	apiMapContent = appendPair(apiMapContent, "versions", nodeForMapping(versionsMapContent))
	artifactsMapContent := nodeSlice()
	_ = core.ListArtifactsForParent(ctx, client, m, func(message *rpc.Artifact) {
		artifactsMapContent = appendPair(artifactsMapContent,
			path.Base(message.Name),
			nodeForMapping(exportArtifact(ctx, client, message)))
	})
	if len(artifactsMapContent) > 0 {
		apiMapContent = appendPair(apiMapContent, "artifacts", nodeForMapping(artifactsMapContent))
	}
	return apiMapContent
}

func exportVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) []*yaml.Node {
	m := names.VersionRegexp().FindStringSubmatch(message.Name)
	versionMapContent := nodeSlice()
	versionMapContent = appendPair(versionMapContent, "createTime", nodeForTime(message.CreateTime.AsTime()))
	versionMapContent = appendPair(versionMapContent, "state", nodeForString(message.State))
	specsMapContent := nodeSlice()
	err := core.ListSpecs(ctx, client, m, "", func(message *rpc.ApiSpec) {
		specMapContent := exportSpec(ctx, client, message)
		specsMapContent = appendPair(specsMapContent, path.Base(message.Name), nodeForMapping(specMapContent))
		m := names.SpecRegexp().FindStringSubmatch(message.Name)
		err := core.ListSpecRevisions(ctx, client, m, "", func(message *rpc.ApiSpec) {
			specMapContent := exportSpec(ctx, client, message)
			specsMapContent = appendPair(specsMapContent, path.Base(message.Name), nodeForMapping(specMapContent))
		})
		check(err)
	})
	check(err)
	versionMapContent = appendPair(versionMapContent, "specs", nodeForMapping(specsMapContent))
	artifactsMapContent := nodeSlice()
	_ = core.ListArtifactsForParent(ctx, client, m, func(message *rpc.Artifact) {
		artifactsMapContent = appendPair(artifactsMapContent,
			path.Base(message.Name),
			nodeForMapping(exportArtifact(ctx, client, message)))
	})
	if len(artifactsMapContent) > 0 {
		versionMapContent = appendPair(versionMapContent, "artifacts", nodeForMapping(artifactsMapContent))
	}
	return versionMapContent
}

func exportSpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec) []*yaml.Node {
	specMapContent := nodeSlice()
	specMapContent = appendPair(specMapContent, "mime_type", nodeForString(message.MimeType))
	specMapContent = appendPair(specMapContent, "hash", nodeForString(message.Hash))
	specMapContent = appendPair(specMapContent, "size", nodeForInt64(int64(message.SizeBytes)))
	specMapContent = appendPair(specMapContent, "createTime", nodeForTime(message.CreateTime.AsTime()))
	specMapContent = appendPair(specMapContent, "revisionId", nodeForString(message.RevisionId))
	return specMapContent
}

func exportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) []*yaml.Node {
	artifactMapContent := nodeSlice()
	artifactMapContent = appendPair(artifactMapContent, "mime_type", nodeForString(message.GetMimeType()))
	artifactMapContent = appendPair(artifactMapContent, "contents", nodeForString(fmt.Sprintf("%+v", message.GetContents())))
	artifactMapContent = appendPair(artifactMapContent, "createTime", nodeForTime(message.CreateTime.AsTime()))
	return artifactMapContent
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

func nodeForString(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
}

func nodeForInt64(value int64) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!int",
		Value: fmt.Sprintf("%d", value),
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
