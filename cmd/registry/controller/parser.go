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
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const resourceKW = "$resource"

func ExtendSourcePattern(
	resourcePattern string,
	sourcePattern string,
	projectID string) (string, error) {
	// Extends the source pattern by replacing references to $resource
	// Example:
	// resourcePattern: "apis/-/versions/-/specs/-/artifacts/-"
	// sourcePattern: "$resource.spec"
	// Returns "apis/-/versions/-/specs/-"

	// resourcePattern: "apis/-/versions/-/specs/-/artifacts/-"
	// sourcePattern: "$resource.api/versions/-"
	// Returns "apis/-/versions/-"

	if !strings.HasPrefix(sourcePattern, resourceKW) {
		return fmt.Sprintf("projects/%s/locations/global/%s", projectID, sourcePattern), nil
	}

	entityRegex := regexp.MustCompile(fmt.Sprintf(`\%s\.(api|version|spec|artifact)`, resourceKW))
	matches := entityRegex.FindStringSubmatch(sourcePattern)
	if len(matches) <= 1 {
		return "", errors.New(fmt.Sprintf("Invalid source pattern: %s", sourcePattern))
	}

	entity, entityType := matches[0], matches[1]
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
		return "", errors.New(fmt.Sprintf("Invalid source pattern: %s", sourcePattern))
	}

	return strings.Replace(sourcePattern, entity, entityVal, 1), nil

}

func ResourceNameFromDependency(
	resourcePattern string,
	dependency Resource) (string, error) {
	// Derives the resource name from the provided resourcePattern and dependencyName.
	// Example:
	// resourcePattern: apis/-/versions/-/specs/-
	// dependencyName: apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity
	// returns apis/petstore/versions/1.0.0/specs/opennapi.yaml
	// Replace apis/- pattern with the corresponding name and so on.
	re := regexp.MustCompile(`.*(/apis/)-`)
	resourceName := string(
		re.ReplaceAll([]byte(resourcePattern), []byte(fmt.Sprintf("%s", dependency.GetApi()))))

	re = regexp.MustCompile(`.*(/versions/)-`)
	resourceName = string(
		re.ReplaceAll([]byte(resourceName), []byte(fmt.Sprintf("%s", dependency.GetVersion()))))

	re = regexp.MustCompile(`.*(/specs/)-`)
	resourceName = string(
		re.ReplaceAll([]byte(resourceName), []byte(fmt.Sprintf("%s", dependency.GetSpec()))))

	re = regexp.MustCompile(`.*(/artifacts/)-`)
	resourceName = string(
		re.ReplaceAll([]byte(resourceName), []byte(fmt.Sprintf("%s", dependency.GetArtifact()))))

	return resourceName, nil

}

func ExtractGroup(pattern string, resource Resource) (string, error) {
	// Reads the sourcePattern, finds out group by entity type and returns the group value
	// Example:
	// pattern: $resource.api/versions/-/specs/-
	// resource: "projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml"
	// returns "projects/demo/apis/petstore"

	if strings.HasPrefix(pattern, resourceKW) {
		// Example:
		// pattern: "$resource.api/versions/-/specs/-"
		// re.FindStringSubmatch will return:
		// ["$resource.api", "api"]
		re := regexp.MustCompile(fmt.Sprintf(`\%s\.(api|version|spec|artifact)`, resourceKW))

		matches := re.FindStringSubmatch(pattern)
		if len(matches) <= 1 {
			return "", errors.New(fmt.Sprintf("Invalid pattern: Cannot extract group from pattern %s", pattern))
		}

		switch entityType := matches[1]; entityType {
		case "api":
			return resource.GetApi(), nil
		case "version":
			return resource.GetVersion(), nil
		case "spec":
			return resource.GetSpec(), nil
		case "artifact":
			return resource.GetArtifact(), nil
		}
	}

	return "default", nil

}

func GenerateCommand(action string, args []Resource) (string, error) {
	// Check if there is a reference to $dep in the action
	isMatch, err := regexp.MatchString(`\$[0-9]`, action)
	if err != nil {
		return "", err
	}
	if !isMatch {
		return action, nil
	}

	for i, resource := range args {
		// Extract the $dep patterns from action
		re := regexp.MustCompile(fmt.Sprintf(`.*(\$%d(\.api|\.version|\.spec|\.artifact)?)`, i))
		// The above func FindStringSubmatch will always return a slice of size 3
		// Example:
		// re.FindStringSubmatch("compute lint $dep0") = ["compute lint $dep0", "$dep0", ""]
		// re.FindStringSubmatch("compute lint $dep0.spec") = ["compute lint $dep0.spec", "$dep0.spec", ".spec"]
		match := re.FindStringSubmatch(action)

		if len(match) == 3 {
			entity, entityType := match[1], match[2]

			entityVal := ""
			if len(entityType) > 0 { // If the reference is present as $dep.api
				switch entityType[1:] {
				case "api":
					entityVal = resource.GetApi()
				case "version":
					entityVal = resource.GetVersion()
				case "spec":
					entityVal = resource.GetSpec()
				case "artifact":
					entityVal = resource.GetArtifact()
				}

				if len(entityVal) == 0 {
					return "", errors.New(fmt.Sprintf("Error generating command, cannot derive args for action. Invalid action: %s", action))
				}
				action = strings.ReplaceAll(action, entity, entityVal)

			} else if len(entity) > 0 { //if only source is present. Eg: $dep0
				entityVal := resource.GetName()
				action = strings.ReplaceAll(action, entity, entityVal)
			}
		} else {
			return "", errors.New(fmt.Sprintf("Error generating command, cannot derive args for action. Invalid action: %s", action))
		}
	}

	return action, nil
}
