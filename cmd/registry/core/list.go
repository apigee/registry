package core

import (
	"context"
	"log"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
)

func ListProjects(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler ProjectHandler) error {
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

func ListAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler ApiHandler) error {
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

func ListVersions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler VersionHandler) error {
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

func ListSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler SpecHandler) error {
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

func ListSpecRevisions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler SpecHandler) error {
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

func ListProperties(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler PropertyHandler) error {
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
		if filter != "" {
			filter += " && "
		}
		filter += "property_id == '" + segments[8] + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	log.Printf("REQUEST %+v\n\n\n", request)
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

func ListPropertiesForParent(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler PropertyHandler) error {
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

func ListLabels(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler LabelHandler) error {
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

func ListLabelsForParent(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler LabelHandler) error {
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
