// Copyright 2021 Google LLC
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

package controller

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

const resourceKW = "$resource"

func parseResourceCollection(resourcePattern string) (ResourceName, error) {
	if api, err := names.ParseApiCollection(resourcePattern); err == nil {
		return ApiName{Api: api}, nil
	} else if version, err := names.ParseVersionCollection(resourcePattern); err == nil {
		return VersionName{Version: version}, nil
	} else if spec, err := names.ParseSpecCollection(resourcePattern); err == nil {
		return SpecName{Spec: spec}, nil
	} else if artifact, err := names.ParseArtifactCollection(resourcePattern); err == nil {
		return ArtifactName{Artifact: artifact}, nil
	}

	return nil, fmt.Errorf("invalid resourcePattern: %s", resourcePattern)
}

func parseResource(resourcePattern string) (ResourceName, error) {
	if api, err := names.ParseApi(resourcePattern); err == nil {
		return ApiName{Api: api}, nil
	} else if version, err := names.ParseVersion(resourcePattern); err == nil {
		return VersionName{Version: version}, nil
	} else if spec, err := names.ParseSpec(resourcePattern); err == nil {
		return SpecName{Spec: spec}, nil
	} else if artifact, err := names.ParseArtifact(resourcePattern); err == nil {
		return ArtifactName{Artifact: artifact}, nil
	}

	return nil, fmt.Errorf("invalid resourcePattern: %s", resourcePattern)
}

func parseResourcePattern(resourcePattern string) (ResourceName, error) {

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

func extendDependencyPattern(
	resourcePattern string,
	dependencyPattern string,
	projectID string) (string, error) {
	// Extends the source pattern by replacing references to $resource
	// Example:
	// resourcePattern: "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-"
	// dependencyPattern: "$resource.spec"
	// Returns "projects/demo/locations/global/apis/-/versions/-/specs/-"

	// resourcePattern: "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-"
	// dependencyPattern: "$resource.api/versions/-"
	// Returns "projects/demo/locations/global/apis/-/versions/-"

	// If there is no $resource prefix, prepend project name and return
	if !strings.HasPrefix(dependencyPattern, resourceKW) {
		return fmt.Sprintf("projects/%s/locations/global/%s", projectID, dependencyPattern), nil
	}

	// Extract the $resource reference
	// Example result for the following regex
	// dependencyPattern: "$resource.api/artifacts/score"
	// matches: ["$resource.api/", "$resource.api", "api"]
	entityRegex := regexp.MustCompile(fmt.Sprintf(`(\%s\.(api|version|spec|artifact))(/|$)`, resourceKW))
	matches := entityRegex.FindStringSubmatch(dependencyPattern)
	if len(matches) <= 2 {
		return "", fmt.Errorf("invalid dependency pattern: %s", dependencyPattern)
	}

	// Convert resourcePattern to resourceName to extract entity values (api, spec, version,artifact)
	resourceName, err := parseResourcePattern(resourcePattern)
	if err != nil {
		return "", err
	}

	entity, entityType := matches[1], matches[2]
	entityVal := ""
	switch entityType {
	case "api":
		entityVal = resourceName.GetApi()
	case "version":
		entityVal = resourceName.GetVersion()
	case "spec":
		entityVal = resourceName.GetSpec()
	case "artifact":
		entityVal = resourceName.GetArtifact()
	default:
		return "", fmt.Errorf("invalid combination resourcePattern: %q dependencyPattern: %q", resourcePattern, dependencyPattern)
	}

	if len(entityVal) == 0 {
		return "", fmt.Errorf("invalid combination resourcePattern: %q dependencyPattern: %q", resourcePattern, dependencyPattern)
	}

	return strings.Replace(dependencyPattern, entity, entityVal, 1), nil

}

func resourceNameFromParent(
	resourcePattern string,
	parent string) (ResourceName, error) {
	// Derives the resource name from the provided resourcePattern and it's parent.
	// Example:
	// 1) resourcePattern: projects/demo/locations/global/apis/-/versions/-/specs/openapi.yaml
	//    parent: projects/demo/locations/global/apis/petstore/versions/1.0.0
	//    returns projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml
	// 2) resourcePattern: projects/demo/locations/global/apis/petstore/versions/-/specs/-/artifacts/custom-artifact
	//    parent: projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml
	//    returns projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/custom-artifact

	parsedResourcePattern, err := parseResourcePattern(resourcePattern)
	if err != nil {
		return nil, fmt.Errorf("Invalid target Pattern: %s", err)
	}

	// Replace the parent pattern in the resourcePattern with the supplied pattern name
	resourceName := strings.Replace(resourcePattern, parsedResourcePattern.GetParent(), parent, 1)

	//Validate generated resourceName
	resource, err := parseResource(resourceName)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %q cannot derive GeneratedResource name from parent %s", resourcePattern, parent)
	}

	return resource, nil
}

func ValidateResourceEntry(resource *rpc.GeneratedResource) error {
	var group string

	for _, dependency := range resource.Dependencies {
		// Validate that all the dependencies are grouped at the same level
		groupEntity, err := getEntityType(dependency.Pattern)
		if err != nil {
			return err
		}
		if len(group) == 0 {
			group = groupEntity
		} else {
			if groupEntity != group {
				return fmt.Errorf("invalid matching: all the dependencies should be matched at the same level from $resource")
			}
		}
	}

	// Validate that the action contains reference to valid entities.
	// Same as the group entity
	entity, entityType, err := getCommandEntity(resource.Action)
	if err != nil {
		return err
	}
	// no $resource reference present
	if len(entity) == 0 {
		return nil
	}
	if entityType != group {
		return fmt.Errorf("invalid reference ($resource.entity) in %q , entity should be same as the match in the dependencies", resource.Action)
	}

	return nil
}

func getEntityType(sourcePattern string) (string, error) {
	// Reads the sourcePattern, finds out entity type in the $resource refernce
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// returns "api", default if no reference is present

	if !strings.HasPrefix(sourcePattern, resourceKW) {
		return "default", nil
	}

	// Example:
	// pattern: "$resource.api/versions/-/specs/-"
	// re.FindStringSubmatch will return:
	// ["$resource.api", "api"]
	re := regexp.MustCompile(fmt.Sprintf(`\%s\.(api|version|spec|artifact)(/|$)`, resourceKW))

	matches := re.FindStringSubmatch(sourcePattern)
	if len(matches) <= 1 {
		return "", fmt.Errorf("invalid pattern: Cannot extract entityType from pattern %s", sourcePattern)
	}

	return matches[1], nil
}

func getEntityKey(sourcePattern string, resource ResourceName) (string, error) {
	// Reads the sourcePattern and returns the entity value for the $resource reference
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// resource: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml"
	// returns "projects/demo/locations/global/apis/petstore"

	entityType, err := getEntityType(sourcePattern)
	if err != nil {
		return "", err
	}

	switch entityType {
	case "api":
		return resource.GetApi(), nil
	case "version":
		return resource.GetVersion(), nil
	case "spec":
		return resource.GetSpec(), nil
	case "artifact":
		return resource.GetArtifact(), nil
	case "default":
		return "default", nil
	default:
		return "", fmt.Errorf("invalid pattern: Cannot extract entity from sourcePattern %s", sourcePattern)
	}

}

func getCommandEntity(action string) (string, string, error) {
	// Check if there is a reference to $n in the action
	isMatch, err := regexp.MatchString(fmt.Sprintf(`\%s`, resourceKW), action)
	if err != nil {
		return "", "", err
	}
	// No $resource references in the action string
	if !isMatch {
		return "", "", nil
	}

	// Extract the $resource patterns from action
	// action = "compute lintstats $resource.spec"
	// This expression will match $resource.spec
	re := regexp.MustCompile(fmt.Sprintf(`\%s(\.api|\.version|\.spec|\.artifact)($|/| )`, resourceKW))
	match := re.FindAllString(action, -1)
	if len(match) == 0 {
		return "", "", fmt.Errorf("invalid action: %s missing or incorrect entity in the reference", action)
	}

	// Check if all the references are at the same level
	entity := strings.TrimRight(match[0], " /")
	for _, m := range match {
		if strings.Trim(m, " /") != entity {
			return "", "", fmt.Errorf("invalid action: %s All the $resource references must be at the same level", action)
		}
	}

	// entity = $resource.api, extract the  entityType as "api"
	entityType := entity[len(resourceKW)+1:]

	return entity, entityType, nil
}

func generateCommand(action string, resourceName string) (string, error) {

	entity, entityType, err := getCommandEntity(action)
	if err != nil {
		return "", err
	}

	// no $resource reference, return the original action
	if len(entity) == 0 {
		return action, nil
	}

	resource, err := parseResource(resourceName)
	if err != nil {
		return "", fmt.Errorf("error generating command, invalid resourceName: %s", resourceName)
	}

	entityVal := ""
	switch entityType {
	case "api":
		entityVal = resource.GetApi()
	case "version":
		entityVal = resource.GetVersion()
	case "spec":
		entityVal = resource.GetSpec()
	case "artifact":
		entityVal = resource.GetArtifact()
	default:
		entityVal = ""
	}

	if len(entityVal) == 0 {
		return "", fmt.Errorf("error generating command, cannot derive args for action. Invalid action: %s", action)
	}
	action = strings.ReplaceAll(action, entity, entityVal)

	return action, nil
}
