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
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/service/registry/names"
	"regexp"
	"strings"
)

const resourceKW = "$resource"

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

	if !strings.HasPrefix(dependencyPattern, resourceKW) {
		return fmt.Sprintf("projects/%s/locations/global/%s", projectID, dependencyPattern), nil
	}

	entityRegex := regexp.MustCompile(fmt.Sprintf(`(\%s\.(api|version|spec|artifact))(/|$)`, resourceKW))
	matches := entityRegex.FindStringSubmatch(dependencyPattern)
	// dependencyPattern: "$resource.api/artifacts/score"
	// matches: ["$resource.api/", "$resource.api", "api"]
	if len(matches) <= 2 {
		return "", fmt.Errorf("invalid dependency pattern: %s", dependencyPattern)
	}

	entity, entityType := matches[1], matches[2]
	entityVal := ""
	switch entityType {
	case "api":
		re := regexp.MustCompile(`.*/apis/[^/]*`)
		entityVal = re.FindString(resourcePattern)
	case "version":
		re := regexp.MustCompile(`.*/versions/[^/]*`)
		entityVal = re.FindString(resourcePattern)
	case "spec":
		re := regexp.MustCompile(`.*/specs/[^/]*`)
		entityVal = re.FindString(resourcePattern)
	case "artifact":
		re := regexp.MustCompile(`.*/artifacts/[^/]*`)
		entityVal = re.FindString(resourcePattern)
	default:
		return "", fmt.Errorf("invalid combination resourcePattern: %q dependencyPattern: %q", resourcePattern, dependencyPattern)
	}

	if len(entityVal) == 0 {
		return "", fmt.Errorf("invalid combination resourcePattern: %q dependencyPattern: %q", resourcePattern, dependencyPattern)
	}

	return strings.Replace(dependencyPattern, entity, entityVal, 1), nil

}

func resourceNameFromGroupKey(
	resourcePattern string,
	groupKey string) (string, error) {
	// Derives the resource name from the provided resourcePattern and dependencyName.
	// Example:
	// 1) resourcePattern: projects/demo/locations/global/apis/-/versions/-/specs/-
	//    dependencyName: projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity
	//    returns projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml
	// 2) resourcePattern: projects/demo/locations/global/apis/petstore/versions/-/specs/-/artifacts/custom-artifact
	//    dependencyName: projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity
	//    returns projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/custom-artifact

	// Replace apis/- pattern with the corresponding name and so on.
	// dependency.GetApi() returns the full api name projects/demo/locations/global/apis/petstore
	// We use stringsSplit()[-1] to extract only the API name
	// apiPattern := regexp.MustCompile(`/apis/-`)
	apiName := strings.Split(extractEntityName(groupKey, "api"), "/")
	resourceName := strings.ReplaceAll(resourcePattern, "/apis/-",
		fmt.Sprintf("/apis/%s", apiName[len(apiName)-1]))

	versionName := strings.Split(extractEntityName(groupKey, "version"), "/")
	resourceName = strings.ReplaceAll(resourceName, "/versions/-",
		fmt.Sprintf("/versions/%s", versionName[len(versionName)-1]))

	specName := strings.Split(extractEntityName(groupKey, "spec"), "/")
	resourceName = strings.ReplaceAll(resourceName, "/specs/-",
		fmt.Sprintf("/specs/%s", specName[len(specName)-1]))

	artifactName := strings.Split(extractEntityName(groupKey, "artifact"), "/")
	resourceName = strings.ReplaceAll(resourceName, "/artifacts/-",
		fmt.Sprintf("/artifacts/%s", artifactName[len(artifactName)-1]))

	//Validate resourceName
	if m := names.ProjectRegexp().FindStringSubmatch(resourceName); m != nil {
		return resourceName, nil
	} else if m := names.ApiRegexp().FindStringSubmatch(resourceName); m != nil {
		return resourceName, nil
	} else if m := names.VersionRegexp().FindStringSubmatch(resourceName); m != nil {
		return resourceName, nil
	} else if m := names.SpecRegexp().FindStringSubmatch(resourceName); m != nil {
		return resourceName, nil
	} else if m := names.ArtifactRegexp().FindStringSubmatch(resourceName); m != nil {
		return resourceName, nil
	}

	return "", fmt.Errorf("invalid pattern: %q cannot derive GeneratedResource name from groupKey %s", resourcePattern, groupKey)
}

func ValidateResourceEntry(resource *rpc.GeneratedResource) error {
	var group string

	for _, dependency := range resource.Dependencies {
		// Validate that all the dependencies are grouped at the same level
		groupEntity, err := getGroupEntity(dependency.Pattern)
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
	entity, entityType, err := getCommandEntitity(resource.Action)
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

func getGroupEntity(pattern string) (string, error) {
	// Reads the sourcePattern, finds out group by entity type
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// returns "api", default if no group is present

	if !strings.HasPrefix(pattern, resourceKW) {
		return "default", nil
	}

	// Example:
	// pattern: "$resource.api/versions/-/specs/-"
	// re.FindStringSubmatch will return:
	// ["$resource.api", "api"]
	re := regexp.MustCompile(fmt.Sprintf(`\%s\.(api|version|spec|artifact)(/|$)`, resourceKW))

	matches := re.FindStringSubmatch(pattern)
	if len(matches) <= 1 {
		return "", fmt.Errorf("invalid pattern: Cannot extract group from pattern %s", pattern)
	}

	return matches[1], nil
}

func getGroupKey(pattern string, resource Resource) (string, error) {
	// Reads the sourcePattern, finds out group by entity type and returns the group value
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// resource: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml"
	// returns "projects/demo/locations/global/apis/petstore"

	entityType, err := getGroupEntity(pattern)
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
		return "", fmt.Errorf("invalid pattern: Cannot extract group from pattern %s", pattern)
	}

}

func getCommandEntitity(action string) (string, string, error) {
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

	entity, entityType, err := getCommandEntitity(action)
	if err != nil {
		return "", err
	}

	// no $resource reference, return the original action
	if len(entity) == 0 {
		return action, nil
	}

	entityVal := extractEntityName(resourceName, entityType)

	if len(entityVal) == 0 {
		return "", fmt.Errorf("error generating command, cannot derive args for action. Invalid action: %s", action)
	}
	action = strings.ReplaceAll(action, entity, entityVal)

	return action, nil
}

func extractEntityName(name string, group_name string) string {
	re := regexp.MustCompile(fmt.Sprintf(".*\\/%ss\\/[^\\/]*", group_name))
	group := re.FindString(name)
	return group
}
