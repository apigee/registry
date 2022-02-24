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

// This function is used in case of receipt artifacts where the name of the target resource to create is not known
func resourceNameFromEntityKey(
	resourcePattern string,
	entityKey string) (string, error) {
	// Derives the resource name from the provided resourcePattern and entityKey.
	// Example:
	// 1) resourcePattern: projects/demo/locations/global/apis/-/versions/-/specs/-
	//    entityKey: projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity
	//    returns projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml
	// 2) resourcePattern: projects/demo/locations/global/apis/petstore/versions/-/specs/-/artifacts/custom-artifact
	//    entityKey: projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity
	//    returns projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/custom-artifact

	entityName, err := parseResource(entityKey)
	resourceName := resourcePattern

	if err == nil {
		// Replace `apis/-` pattern with the corresponding name.
		// groupName.GetApi() returns the full api name projects/demo/locations/global/apis/petstore
		// We use stringsSplit()[-1] to extract only the API name
		apiName := strings.Split(entityName.GetApi(), "/")
		if len(apiName) > 0 {
			resourceName = strings.ReplaceAll(resourceName, "/apis/-",
				fmt.Sprintf("/apis/%s", apiName[len(apiName)-1]))
		}

		versionName := strings.Split(entityName.GetVersion(), "/")
		if len(versionName) > 0 {
			resourceName = strings.ReplaceAll(resourceName, "/versions/-",
				fmt.Sprintf("/versions/%s", versionName[len(versionName)-1]))
		}

		specName := strings.Split(entityName.GetSpec(), "/")
		if len(specName) > 0 {
			resourceName = strings.ReplaceAll(resourceName, "/specs/-",
				fmt.Sprintf("/specs/%s", specName[len(specName)-1]))
		}

		artifactName := strings.Split(entityName.GetArtifact(), "/")
		if len(artifactName) > 0 {
			resourceName = strings.ReplaceAll(resourceName, "/artifacts/-",
				fmt.Sprintf("/artifacts/%s", artifactName[len(artifactName)-1]))
		}
	}
	//Validate generated resourceName
	if _, err := names.ParseProject(resourceName); err == nil {
		return resourceName, nil
	} else if _, err := names.ParseApi(resourceName); err == nil {
		return resourceName, nil
	} else if _, err := names.ParseVersion(resourceName); err == nil {
		return resourceName, nil
	} else if _, err := names.ParseSpec(resourceName); err == nil {
		return resourceName, nil
	} else if _, err := names.ParseArtifact(resourceName); err == nil {
		return resourceName, nil
	}

	return "", fmt.Errorf("invalid pattern: %q cannot derive GeneratedResource name from entityKey %s", resourcePattern, entityKey)
}

func compareEntities(e1, e2 string) int {
	order := map[string]int{
		"artifact": 0,
		"spec":     1,
		"version":  2,
		"api":      3,
		"":         4,
	}

	return order[e1] - order[e2]
}

func minRefIndexFromDependency(generatedResource *rpc.GeneratedResource) (string, int, error) {
	// Find the lowest entity reference in the dependencies
	// ""(no $resource) > $resource.api > $resource.version > $resource.spec > $resource.artifact
	minDepEntityType := ""
	minIndex := -1
	for i, dependency := range generatedResource.Dependencies {
		extractedEntity, err := getReferenceFromDependency(dependency.Pattern)
		if err != nil {
			return "", -1, err
		}
		if compareEntities(extractedEntity, minDepEntityType) <= 0 {
			minDepEntityType = extractedEntity
			minIndex = i
		}
	}

	return minDepEntityType, minIndex, nil
}

func ValidateResourceEntry(resource *rpc.GeneratedResource) error {

	// Find the lowest entity reference in the dependencies
	// ""(no $resource) > $resource.api > $resource.version > $resource.spec > $resource.artifact
	minDepEntityType, _, err := minRefIndexFromDependency(resource)
	if err != nil {
		return err
	}

	// Find the lowest entity reference in the action string
	// ""(no $resource) > $resource.api > $resource.version > $resource.spec > $resource.artifact
	references, err := getReferencesFromAction(resource.Action)
	if err != nil {
		return err
	}
	// no $resource reference present
	if len(references) == 0 {
		return nil
	}

	minActionEntityType := ""
	for _, r := range references {
		if compareEntities(r.entityType, minActionEntityType) < 0 {
			minActionEntityType = r.entityType
		}
	}

	// Check if minDepEntityType and minActionEntityType are same
	if minDepEntityType != minActionEntityType {
		return fmt.Errorf("Invalid $resource references: The The lowest entity reference %q in dependencies is not same as the lowest entity reference in the action string %q.",
			minDepEntityType, minActionEntityType)
	}

	return nil
}

func getReferenceFromDependency(pattern string) (string, error) {
	// Reads the sourcePattern, finds out entity type in the reference
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// returns "api", "" if no group is present

	if !strings.HasPrefix(pattern, resourceKW) {
		return "", nil
	}

	// Example:
	// pattern: "$resource.api/versions/-/specs/-"
	// re.FindStringSubmatch will return:
	// ["$resource.api", "api"]
	re := regexp.MustCompile(fmt.Sprintf(`\%s\.(api|version|spec|artifact)(/|$)`, resourceKW))

	matches := re.FindStringSubmatch(pattern)
	if len(matches) <= 1 {
		return "", fmt.Errorf("invalid pattern: Cannot extract referenced entity from pattern %s", pattern)
	}

	return matches[1], nil
}

func getEntityKey(pattern string, resource ResourceInstance) (string, error) {
	// Reads the pattern and returns the entity value for the resource
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// resource: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml"
	// returns "projects/demo/locations/global/apis/petstore"

	entityType, err := getReferenceFromDependency(pattern)
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
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("invalid pattern: Cannot extract entity from pattern %s", pattern)
	}

}

type reference struct {
	entity     string
	entityType string
}

func getReferencesFromAction(action string) ([]reference, error) {
	references := make([]reference, 0)

	// Check if there is a reference to $resource in the action
	isMatch, err := regexp.MatchString(fmt.Sprintf(`\%s`, resourceKW), action)
	if err != nil {
		return nil, err
	}
	// No $resource references in the action string
	if !isMatch {
		return references, nil
	}

	// Extract the $resource patterns from action
	// action = "compute lintstats $resource.spec"
	// This expression will match $resource.spec
	re := regexp.MustCompile(fmt.Sprintf(`\%s(\.api|\.version|\.spec|\.artifact)($|/| )`, resourceKW))
	match := re.FindAllString(action, -1)
	if len(match) == 0 {
		return nil, fmt.Errorf("invalid action: %s missing or incorrect entity in the reference", action)
	}

	// Construct a list of entity: entityType for all the references
	for _, m := range match {
		// entity = $resource.api, extract the entityType as "api"
		entity := strings.TrimRight(m, " /")
		entityType := entity[len(resourceKW)+1:]
		references = append(references, reference{entity: entity, entityType: entityType})
	}

	return references, nil
}

func generateCommand(action string, resourceName string) (string, error) {

	references, err := getReferencesFromAction(action)
	if err != nil {
		return "", err
	}

	// no $resource reference, return the original action
	if len(references) == 0 {
		return action, nil
	}

	resource, err := parseResource(resourceName)
	if err != nil {
		return "", fmt.Errorf("error generating command, invalid resourceName: %s", resourceName)
	}

	for _, r := range references {
		entityVal := ""
		switch r.entityType {
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
		action = strings.ReplaceAll(action, r.entity, entityVal)
	}

	return action, nil
}
