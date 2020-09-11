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
	"context"
	"fmt"
	"log"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/status"
)

var filterFlag string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&filterFlag, "filter", "", "Filter option to send with list calls")
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources in the Registry.",
	Long:  "List resources in the Registry.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		name := args[0]
		if m := names.ProjectsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listProjects(ctx, client, m[0], printProject)
		} else if m := names.ApisRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listAPIs(ctx, client, m[0], printAPI)
		} else if m := names.VersionsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listVersions(ctx, client, m[0], printVersion)
		} else if m := names.SpecsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listSpecs(ctx, client, m[0], printSpec)
		} else if m := names.PropertiesRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listProperties(ctx, client, m[0], printProperty)
		} else if m := names.LabelsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			err = listLabels(ctx, client, m[0], printLabel)

		} else if m := names.ProjectRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listProjects(ctx, client, segments, printProject)
			} else {
				_, err = getProject(ctx, client, segments, printProject)
			}
		} else if m := names.ApiRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listAPIs(ctx, client, segments, printAPI)
			} else {
				_, err = getAPI(ctx, client, segments, printAPI)
			}
		} else if m := names.VersionRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listVersions(ctx, client, segments, printVersion)
			} else {
				_, err = getVersion(ctx, client, segments, printVersion)
			}
		} else if m := names.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listSpecs(ctx, client, segments, printSpec)
			} else {
				_, err = getSpec(ctx, client, segments, false, printSpec)
			}
		} else if m := names.PropertyRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listProperties(ctx, client, segments, printProperty)
			} else {
				_, err = getProperty(ctx, client, segments, printProperty)
			}
		} else if m := names.LabelRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				err = listLabels(ctx, client, segments, printLabel)
			} else {
				_, err = getLabel(ctx, client, segments, printLabel)
			}
		} else {
			fmt.Printf("unsupported argument(s): %+v\n", args)
		}
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				log.Printf("%s", err.Error())
			} else {
				log.Printf("%s", st.Message())
			}
		}
	},
}

type projectHandler func(*rpc.Project)
type apiHandler func(*rpc.Api)
type versionHandler func(*rpc.Version)
type specHandler func(*rpc.Spec)
type propertyHandler func(*rpc.Property)
type labelHandler func(*rpc.Label)

func printProject(project *rpc.Project) {
	fmt.Println(project.Name)
}

func printAPI(api *rpc.Api) {
	fmt.Println(api.Name)
}

func printVersion(version *rpc.Version) {
	fmt.Println(version.Name)
}

func printSpec(spec *rpc.Spec) {
	fmt.Println(spec.Name)
}

func printProperty(property *rpc.Property) {
	fmt.Println(property.Name)
}

func printLabel(label *rpc.Label) {
	fmt.Println(label.Name)
}

func listProjects(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler projectHandler) error {
	request := &rpc.ListProjectsRequest{}
	filter := filterFlag
	if len(segments) == 2 && segments[1] != "-" {
		filter = "project_id == '" + segments[1] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListProjects(ctx, request)
	for {
		project, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(project)
	}
	return nil
}

func listAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler apiHandler) error {
	request := &rpc.ListApisRequest{
		Parent: "projects/" + segments[1],
	}
	filter := filterFlag
	if len(segments) == 3 && segments[2] != "-" {
		filter = "api_id == '" + segments[2] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListApis(ctx, request)
	for {
		api, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(api)
	}
	return nil
}

func listVersions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler versionHandler) error {
	request := &rpc.ListVersionsRequest{
		Parent: "projects/" + segments[1] + "/apis/" + segments[2],
	}
	filter := filterFlag
	if len(segments) == 4 && segments[3] != "-" {
		filter = "version_id == '" + segments[3] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListVersions(ctx, request)
	for {
		version, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(version)
	}
	return nil
}

func listSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler specHandler) error {
	request := &rpc.ListSpecsRequest{
		Parent: "projects/" + segments[1] + "/apis/" + segments[2] + "/versions/" + segments[3],
	}
	filter := filterFlag
	if len(segments) > 4 && segments[4] != "-" {
		filter = "spec_id == '" + segments[4] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListSpecs(ctx, request)
	for {
		spec, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(spec)
	}
	return nil
}

func listSpecRevisions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler specHandler) error {
	request := &rpc.ListSpecRevisionsRequest{
		Name: "projects/" + segments[1] +
			"/apis/" + segments[2] +
			"/versions/" + segments[3] +
			"/specs/" + segments[4],
	}
	it := client.ListSpecRevisions(ctx, request)
	for {
		spec, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(spec)
	}
	return nil
}

func listProperties(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler propertyHandler) error {
	parent := "projects/" + segments[1]
	if segments[3] != "" {
		parent += "/apis/" + segments[3]
		if segments[5] != "" {
			parent += "/versions/" + segments[5]
			if segments[7] != "" {
				parent += "/specs/" + segments[7]
			}
		}
	}
	request := &rpc.ListPropertiesRequest{
		Parent: parent,
	}
	filter := filterFlag
	if len(segments) == 9 && segments[8] != "-" {
		filter = "property_id == '" + segments[8] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListProperties(ctx, request)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(property)
	}
	return nil
}

func listPropertiesForParent(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler propertyHandler) error {
	parent := "projects/" + segments[1]
	if len(segments) > 2 {
		parent += "/apis/" + segments[2]
		if len(segments) > 3 {
			parent += "/versions/" + segments[3]
			if len(segments) > 4 {
				parent += "/specs/" + segments[4]
			}
		}
	}
	request := &rpc.ListPropertiesRequest{
		Parent: parent,
	}
	it := client.ListProperties(ctx, request)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(property)
	}
	return nil
}

func listLabels(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler labelHandler) error {
	parent := "projects/" + segments[1]
	if segments[3] != "" {
		parent += "/apis/" + segments[3]
		if segments[5] != "" {
			parent += "/versions/" + segments[5]
			if segments[7] != "" {
				parent += "/specs/" + segments[7]
			}
		}
	}
	request := &rpc.ListLabelsRequest{
		Parent: parent,
	}
	filter := filterFlag
	if len(segments) == 9 && segments[8] != "-" {
		filter = "label_id == '" + segments[8] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListLabels(ctx, request)
	for {
		label, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(label)
	}
	return nil
}

func listLabelsForParent(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler labelHandler) error {
	parent := "projects/" + segments[1]
	if len(segments) > 2 {
		parent += "/apis/" + segments[2]
		if len(segments) > 3 {
			parent += "/versions/" + segments[3]
			if len(segments) > 4 {
				parent += "/specs/" + segments[4]
			}
		}
	}
	request := &rpc.ListLabelsRequest{
		Parent: parent,
	}
	it := client.ListLabels(ctx, request)
	for {
		label, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(label)
	}
	return nil
}

func getProject(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler projectHandler) (*rpc.Project, error) {
	request := &rpc.GetProjectRequest{
		Name: "projects/" + segments[1],
	}
	project, err := client.GetProject(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(project)
	}
	return project, nil
}

func getAPI(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler apiHandler) (*rpc.Api, error) {
	request := &rpc.GetApiRequest{
		Name: "projects/" + segments[1] + "/apis/" + segments[2],
	}
	api, err := client.GetApi(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(api)
	}
	return api, nil
}

func getVersion(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler versionHandler) (*rpc.Version, error) {
	request := &rpc.GetVersionRequest{
		Name: "projects/" + segments[1] + "/apis/" + segments[2] + "/versions/" + segments[3],
	}
	version, err := client.GetVersion(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(version)
	}
	return version, nil
}

func getSpec(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	getContents bool,
	handler specHandler) (*rpc.Spec, error) {
	view := rpc.SpecView_BASIC
	if getContents {
		view = rpc.SpecView_FULL
	}
	request := &rpc.GetSpecRequest{
		Name: "projects/" + segments[1] + "/apis/" + segments[2] + "/versions/" + segments[3] + "/specs/" + segments[4],
		View: view,
	}
	spec, err := client.GetSpec(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(spec)
	}
	return spec, nil
}

func getProperty(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler propertyHandler) (*rpc.Property, error) {
	request := &rpc.GetPropertyRequest{
		// TODO: fix for properties on other resources (besides specs)
		Name: "projects/" + segments[1] + "/apis/" + segments[3] + "/versions/" + segments[5] + "/specs/" + segments[7] + "/properties/" + segments[8],
	}
	property, err := client.GetProperty(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(property)
	}
	return property, nil
}

func getLabel(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler labelHandler) (*rpc.Label, error) {
	request := &rpc.GetLabelRequest{
		// TODO: fix for labels on other resources (besides projects)
		Name: "projects/" + segments[1] + "/labels/" + segments[2],
	}
	label, err := client.GetLabel(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(label)
	}
	return label, nil
}

func sliceContainsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
