// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package patterns

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/apigee/registry/pkg/names"
)

const ResourceKW = "$resource"

func parseResourceCollection(resourcePattern string) (ResourceName, error) {
	if project, err := names.ParseProjectCollection(resourcePattern); err == nil {
		return ProjectName{Name: project}, nil
	} else if api, err := names.ParseApiCollection(resourcePattern); err == nil {
		return ApiName{Name: api}, nil
	} else if version, err := names.ParseVersionCollection(resourcePattern); err == nil {
		return VersionName{Name: version}, nil
	} else if spec, err := names.ParseSpecCollection(resourcePattern); err == nil {
		return SpecName{Name: spec}, nil
	} else if artifact, err := names.ParseArtifactCollection(resourcePattern); err == nil {
		return ArtifactName{Name: artifact}, nil
	}

	return nil, fmt.Errorf("invalid resourcePattern: %s", resourcePattern)
}

func parseResource(resourcePattern string) (ResourceName, error) {
	if project, err := names.ParseProject(resourcePattern); err == nil {
		return ProjectName{Name: project}, nil
	} else if project, err := names.ParseProjectWithLocation(resourcePattern); err == nil {
		return ProjectName{Name: project}, nil
	} else if api, err := names.ParseApi(resourcePattern); err == nil {
		return ApiName{Name: api}, nil
	} else if version, err := names.ParseVersion(resourcePattern); err == nil {
		return VersionName{Name: version}, nil
	} else if spec, err := names.ParseSpecRevision(resourcePattern); err == nil {
		return SpecName{Name: spec.Spec(), RevisionID: spec.RevisionID}, nil
	} else if artifact, err := names.ParseArtifact(resourcePattern); err == nil {
		return ArtifactName{Name: artifact}, nil
	}

	return nil, fmt.Errorf("invalid resourcePattern: %s", resourcePattern)
}

func ParseResourcePattern(resourcePattern string) (ResourceName, error) {
	// First try to match resource collections.
	resource, err := parseResourceCollection(resourcePattern)
	if err == nil {
		return resource, nil
	}

	// Then try to match resource names.
	resource, err = parseResource(resourcePattern)
	if err == nil {
		return resource, nil
	}

	return nil, fmt.Errorf("invalid resourcePattern: %s", resourcePattern)
}

func SubstituteReferenceEntity(resourcePattern string, referred ResourceName) (ResourceName, error) {
	// Extends the resource pattern by replacing references to $resource
	// Example:
	// referred: "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-"
	// resourcePattern: "$resource.spec"
	// Returns "projects/demo/locations/global/apis/-/versions/-/specs/-"

	// referred: "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-"
	// resourcePattern: "$resource.api/versions/-"
	// Returns "projects/demo/locations/global/apis/-/versions/-"

	entity, entityType, err := GetReferenceEntityType(resourcePattern)
	if err != nil {
		return nil, err
	}

	// no $resource reference present
	// simply prepend the projectname and return full resource name
	if entityType == "default" {
		resourceName, err := ParseResourcePattern(fmt.Sprintf("%s/locations/global/%s", referred.Project(), resourcePattern))
		if err != nil {
			return nil, err
		}
		return resourceName, nil
	}

	entityVal, err := GetReferenceEntityValue(resourcePattern, referred)
	if err != nil {
		return nil, err
	}

	extendedName, err := ParseResourcePattern(strings.Replace(resourcePattern, entity, entityVal, 1))
	if err != nil {
		return nil, err
	}
	return extendedName, nil
}

func GetReferenceEntityType(resourcePattern string) (entity, entityType string, err error) {
	// Reads the resourcePattern, finds out entity type in the $resource reference
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// returns "$resource.api","api"
	// return "", "default" if no reference is present

	if !strings.HasPrefix(resourcePattern, ResourceKW) {
		entity, entityType = "", "default"
		err = nil
		return
	}

	// Extract the $resource reference
	// Example result for the following regex
	// dependencyPattern: "$resource.api/artifacts/score"
	// matches: ["$resource.api/", "$resource.api", "api"]
	entityRegex := regexp.MustCompile(fmt.Sprintf(`(\%s\.(api|version|spec|artifact))(/|$)`, ResourceKW))
	matches := entityRegex.FindStringSubmatch(resourcePattern)
	if len(matches) <= 2 {
		entity, entityType = "", ""
		err = fmt.Errorf("invalid resourcePattern: %s", resourcePattern)
		return
	}

	entity, entityType = matches[1], matches[2]
	err = nil
	return
}

func GetReferenceEntityValue(resourcePattern string, referred ResourceName) (string, error) {
	// Reads the sourcePattern and returns the entity value for the $resource reference
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// resource: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"
	// returns "projects/demo/locations/global/apis/petstore"

	_, entityType, err := GetReferenceEntityType(resourcePattern)
	if err != nil {
		return "", err
	}

	switch entityType {
	case "api":
		entityVal := referred.Api()
		if len(entityVal) == 0 {
			return "", fmt.Errorf("invalid combination referred: %q resourcePattern: %q", referred, resourcePattern)
		}
		return entityVal, nil
	case "version":
		entityVal := referred.Version()
		if len(entityVal) == 0 {
			return "", fmt.Errorf("invalid combination referred: %q resourcePattern: %q", referred, resourcePattern)
		}
		return entityVal, nil
	case "spec":
		entityVal := referred.Spec()
		if len(entityVal) == 0 {
			return "", fmt.Errorf("invalid combination referred: %q resourcePattern: %q", referred, resourcePattern)
		}
		return entityVal, nil
	case "artifact":
		entityVal := referred.Artifact()
		if len(entityVal) == 0 {
			return "", fmt.Errorf("invalid combination referred: %q resourcePattern: %q", referred, resourcePattern)
		}
		return entityVal, nil
	case "default":
		return "default", nil
	default:
		return "", fmt.Errorf("invalid combination referred: %q resourcePattern: %q", referred, resourcePattern)
	}
}

func FullResourceNameFromParent(resourcePattern string, parent string) (ResourceName, error) {
	// Derives the resource name from the provided resourcePattern and it's parent.
	// Example:
	// 1) resourcePattern: projects/demo/locations/global/apis/-/versions/-/specs/openapi
	//    parent: projects/demo/locations/global/apis/petstore/versions/1.0.0
	//    returns projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi
	// 2) resourcePattern: projects/demo/locations/global/apis/petstore/versions/-/specs/-/artifacts/custom-artifact
	//    parent: projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi
	//    returns projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/custom-artifact

	parsedResourcePattern, err := ParseResourcePattern(resourcePattern)
	if err != nil {
		return nil, fmt.Errorf("invalid target Pattern: %s", err)
	}

	// Replace the parent pattern in the resourcePattern with the supplied pattern name
	resourceName := strings.Replace(resourcePattern, parsedResourcePattern.ParentName().String(), parent, 1)

	//Validate generated resourceName
	resource, err := parseResource(resourceName)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %q cannot derive GeneratedResource name from parent %s", resourcePattern, parent)
	}

	return resource, nil
}
