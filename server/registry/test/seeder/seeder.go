// Copyright 2021 Google LLC. All Rights Reserved.
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

package seeder

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Registry is an interface containing methods necessary for seeding Registry resources.
type Registry interface {
	CreateProject(context.Context, *rpc.CreateProjectRequest) (*rpc.Project, error)
	CreateApi(context.Context, *rpc.CreateApiRequest) (*rpc.Api, error)
	CreateApiVersion(context.Context, *rpc.CreateApiVersionRequest) (*rpc.ApiVersion, error)
	UpdateApiSpec(context.Context, *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error)
	UpdateApiDeployment(context.Context, *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error)
	CreateArtifact(context.Context, *rpc.CreateArtifactRequest) (*rpc.Artifact, error)
}

// RegistryResource is an interface that any seedable resource will implement.
// The resource name is used to determine the hierarchical ordering of resources for seeding.
type RegistryResource interface {
	GetName() string
}

// SeedRegistry initializes registry with the provided resources.
// Resources are created implicitly if they are needed but aren't explicitly provided.
//
// ApiSpecs with the same name can be provided to create multiple revisions of the same spec.
//
// Resource names must be unique with the exception of ApiSpec resources.
// Supported resource types are Project, Api, ApiVersion, ApiSpec, ApiDeployment, and Artifact.
func SeedRegistry(ctx context.Context, s Registry, resources ...RegistryResource) error {
	// Maintain a history of created resources to skip redundant requests.
	h := make(map[string]bool, 5*len(resources))
	for _, resource := range resources {
		switch r := resource.(type) {
		case *rpc.Project:
			if h[r.GetName()] {
				return fmt.Errorf("cannot seed multiple projects with name %s", r.GetName())
			}

			if err := seedProject(ctx, s, r, h); err != nil {
				return err
			}
		case *rpc.Api:
			if h[r.GetName()] {
				return fmt.Errorf("cannot seed multiple apis with name %s", r.GetName())
			}

			if err := seedApi(ctx, s, r, h); err != nil {
				return err
			}
		case *rpc.ApiVersion:
			if h[r.GetName()] {
				return fmt.Errorf("cannot seed multiple versions with name %s", r.GetName())
			}

			if err := seedVersion(ctx, s, r, h); err != nil {
				return err
			}
		case *rpc.ApiSpec:
			if err := seedSpec(ctx, s, r, h); err != nil {
				return err
			}
		case *rpc.ApiDeployment:
			if h[r.GetName()] {
				return fmt.Errorf("cannot seed multiple deployments with name %s", r.GetName())
			}

			if err := seedDeployment(ctx, s, r, h); err != nil {
				return err
			}
		case *rpc.Artifact:
			if h[r.GetName()] {
				return fmt.Errorf("cannot seed multiple artifacts with name %s", r.GetName())
			}

			if err := seedArtifact(ctx, s, r, h); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported resource type %T", r)
		}
	}

	return nil
}

// SeedProjects is a convenience function for calling SeedRegistry with only Project messages.
func SeedProjects(ctx context.Context, s Registry, projects ...*rpc.Project) error {
	resources := make([]RegistryResource, 0, len(projects))
	for _, r := range projects {
		if r != nil {
			resources = append(resources, r)
		}
	}
	return SeedRegistry(ctx, s, resources...)
}

// SeedApis is a convenience function for calling SeedRegistry with only Api messages.
func SeedApis(ctx context.Context, s Registry, apis ...*rpc.Api) error {
	resources := make([]RegistryResource, 0, len(apis))
	for _, r := range apis {
		if r != nil {
			resources = append(resources, r)
		}
	}
	return SeedRegistry(ctx, s, resources...)
}

// SeedVersions is a convenience function for calling SeedRegistry with only ApiVersion messages.
func SeedVersions(ctx context.Context, s Registry, versions ...*rpc.ApiVersion) error {
	resources := make([]RegistryResource, 0, len(versions))
	for _, r := range versions {
		if r != nil {
			resources = append(resources, r)
		}
	}
	return SeedRegistry(ctx, s, resources...)
}

// SeedSpecs is a convenience function for calling SeedRegistry with only ApiSpec messages.
func SeedSpecs(ctx context.Context, s Registry, specs ...*rpc.ApiSpec) error {
	resources := make([]RegistryResource, 0, len(specs))
	for _, r := range specs {
		if r != nil {
			resources = append(resources, r)
		}
	}
	return SeedRegistry(ctx, s, resources...)
}

// SeedDeployments is a convenience function for calling SeedRegistry with only ApiDeployment messages.
func SeedDeployments(ctx context.Context, s Registry, deployments ...*rpc.ApiDeployment) error {
	resources := make([]RegistryResource, 0, len(deployments))
	for _, r := range deployments {
		if r != nil {
			resources = append(resources, r)
		}
	}
	return SeedRegistry(ctx, s, resources...)
}

// SeedArtifacts is a convenience function for calling SeedRegistry with only Artifact messages.
func SeedArtifacts(ctx context.Context, s Registry, artifacts ...*rpc.Artifact) error {
	resources := make([]RegistryResource, 0, len(artifacts))
	for _, r := range artifacts {
		if r != nil {
			resources = append(resources, r)
		}
	}
	return SeedRegistry(ctx, s, resources...)
}

func seedProject(ctx context.Context, s Registry, p *rpc.Project, history map[string]bool) error {
	if id := p.GetName(); history[id] {
		return nil
	} else {
		history[id] = true
	}

	name, err := names.ParseProject(p.GetName())
	if err != nil {
		return err
	}

	req := &rpc.CreateProjectRequest{
		ProjectId: name.ProjectID,
		Project:   p,
	}

	_, err = s.CreateProject(ctx, req)
	return err
}

func seedApi(ctx context.Context, s Registry, api *rpc.Api, history map[string]bool) error {
	if id := api.GetName(); history[id] {
		return nil
	} else {
		history[id] = true
	}

	name, err := names.ParseApi(api.GetName())
	if err != nil {
		return err
	}

	if err := seedProject(ctx, s, &rpc.Project{Name: fmt.Sprintf("projects/%s", name.ProjectID)}, history); err != nil {
		return err
	}

	_, err = s.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: name.Parent(),
		ApiId:  name.ApiID,
		Api:    api,
	})

	return err
}

func seedVersion(ctx context.Context, s Registry, v *rpc.ApiVersion, history map[string]bool) error {
	if name := v.GetName(); history[name] {
		return nil
	} else {
		history[name] = true
	}

	name, err := names.ParseVersion(v.GetName())
	if err != nil {
		return err
	}

	if err := seedApi(ctx, s, &rpc.Api{Name: name.Parent()}, history); err != nil {
		return err
	}

	_, err = s.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		Parent:       name.Parent(),
		ApiVersionId: name.VersionID,
		ApiVersion:   v,
	})

	return err
}

func seedSpec(ctx context.Context, s Registry, spec *rpc.ApiSpec, history map[string]bool) error {
	if id := fmt.Sprintf("%s@%s", spec.GetName(), sha256hash(spec.GetContents())); history[id] {
		return nil
	} else {
		history[id] = true
	}

	name, err := names.ParseSpec(spec.GetName())
	if err != nil {
		return err
	}

	if err := seedVersion(ctx, s, &rpc.ApiVersion{Name: name.Parent()}, history); err != nil {
		return err
	}

	if _, err := s.UpdateApiSpec(ctx, &rpc.UpdateApiSpecRequest{
		ApiSpec:      spec,
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"*"}},
		AllowMissing: true,
	}); err != nil {
		return err
	}

	return nil
}

func sha256hash(bytes []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(bytes))
}

func seedDeployment(ctx context.Context, s Registry, deployment *rpc.ApiDeployment, history map[string]bool) error {
	signature := sha256hash([]byte(fmt.Sprintf("%s:%s", deployment.GetApiSpecRevision(), deployment.GetEndpointUri())))
	if id := fmt.Sprintf("%s@%s", deployment.GetName(), signature); history[id] {
		return nil
	} else {
		history[id] = true
	}

	name, err := names.ParseDeployment(deployment.GetName())
	if err != nil {
		return err
	}

	if err := seedApi(ctx, s, &rpc.Api{Name: name.Parent()}, history); err != nil {
		return err
	}

	if _, err := s.UpdateApiDeployment(ctx, &rpc.UpdateApiDeploymentRequest{
		ApiDeployment: deployment,
		UpdateMask:    &fieldmaskpb.FieldMask{Paths: []string{"*"}},
		AllowMissing:  true,
	}); err != nil {
		return err
	}

	return nil
}

func seedArtifact(ctx context.Context, s Registry, a *rpc.Artifact, history map[string]bool) error {
	if id := a.GetName(); history[id] {
		return nil
	} else {
		history[id] = true
	}

	name, err := names.ParseArtifact(a.GetName())
	if err != nil {
		return err
	}

	if name.SpecID() != "" {
		err = seedSpec(ctx, s, &rpc.ApiSpec{Name: name.Parent()}, history)
	} else if name.VersionID() != "" {
		err = seedVersion(ctx, s, &rpc.ApiVersion{Name: name.Parent()}, history)
	} else if name.DeploymentID() != "" {
		err = seedDeployment(ctx, s, &rpc.ApiDeployment{Name: name.Parent()}, history)
	} else if name.ApiID() != "" {
		err = seedApi(ctx, s, &rpc.Api{Name: name.Parent()}, history)
	} else if name.ProjectID() != "" {
		err = seedProject(ctx, s, &rpc.Project{Name: fmt.Sprintf("projects/%s", name.ProjectID())}, history)
	}

	if err != nil {
		return err
	}

	_, err = s.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
		Parent:     name.Parent(),
		ArtifactId: name.ArtifactID(),
		Artifact:   a,
	})

	return err
}
