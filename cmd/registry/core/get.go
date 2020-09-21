package core

import (
	"context"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
)

func GetProject(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler ProjectHandler) (*rpc.Project, error) {
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

func GetAPI(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler ApiHandler) (*rpc.Api, error) {
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

func GetVersion(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler VersionHandler) (*rpc.Version, error) {
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

func GetSpec(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	getContents bool,
	handler SpecHandler) (*rpc.Spec, error) {
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

func GetProperty(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler PropertyHandler) (*rpc.Property, error) {
	request := &rpc.GetPropertyRequest{}
	if segments[3] == "" {
		request.Name = "projects/" + segments[1]
	} else if segments[5] == "" {
		request.Name = "projects/" + segments[1] + "/apis/" + segments[3]
	} else if segments[7] == "" {
		request.Name = "projects/" + segments[1] + "/apis/" + segments[3] + "/versions/" + segments[5]
	} else {
		request.Name = "projects/" + segments[1] + "/apis/" + segments[3] + "/versions/" + segments[5] + "/specs/" + segments[7]
	}
	request.Name += "/properties/" + segments[8]

	property, err := client.GetProperty(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(property)
	}
	return property, nil
}

func GetLabel(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler LabelHandler) (*rpc.Label, error) {
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
