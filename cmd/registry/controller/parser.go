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

	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/pkg/application/controller"
)

func ValidateManifest(parent string, manifest *controller.Manifest) []error {
	totalErrors := make([]error, 0)
	for _, resource := range manifest.GeneratedResources {
		errs := validateGeneratedResourceEntry(parent, resource)
		for _, err := range errs {
			totalErrors = append(totalErrors, fmt.Errorf("invalid entry: %v, %s", resource, err))
		}
	}
	return totalErrors
}

func validateGeneratedResourceEntry(parent string, generatedResource *controller.GeneratedResource) []error {
	parsedTargetResource, err := patterns.ParseResourcePattern(
		fmt.Sprintf("%s/%s", parent, generatedResource.Pattern))

	// Check that the target resource pattern should be a valid pattern.
	// Return for errors in target resource pattern since we can't verify action and dependencies based off an incorrect pattern.
	if err != nil {
		return []error{
			fmt.Errorf("invalid pattern for generatedResource %v, %s", generatedResource.Pattern, err),
		}
	}
	// Check that generatedResource pattern doesn't end with a "-".
	// We require a name for the target resource.
	if strings.HasSuffix(parsedTargetResource.String(), "/-") {
		return []error{
			fmt.Errorf("invalid generatedResource pattern: %q, it should end with a name and not a \"-\"", generatedResource.Pattern),
		}
	}

	validateEntityReference := func(resourceName patterns.ResourceName, entityType string) bool {
		switch entityType {
		case "api":
			return resourceName.Api() != ""
		case "version":
			return resourceName.Version() != ""
		case "spec":
			return resourceName.Spec() != ""
		case "artifact":
			return resourceName.Artifact() != ""
		case "default":
			return true
		default:
			return false
		}
	}

	errs := make([]error, 0)
	for _, dependency := range generatedResource.Dependencies {
		// Validate that all the dependencies have valid $resource references.
		_, entityType, err := patterns.GetReferenceEntityType(dependency.Pattern)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid dependency pattern %s: %s", dependency.Pattern, err))
		}

		if !validateEntityReference(parsedTargetResource, entityType) {
			errs = append(errs, fmt.Errorf("invalid reference in dependency pattern: %s", dependency.Pattern))
		}
	}

	// Check that either "dependencies" or "refresh" is set and "refresh > 0"
	if len(generatedResource.Dependencies) == 0 && generatedResource.Refresh == nil {
		errs = append(errs, fmt.Errorf("either 'dependencies' or 'refresh' must be set for generated resource: %v", generatedResource))
	}

	// Check that "refresh" > 0
	if generatedResource.Refresh != nil && generatedResource.Refresh.AsDuration().Seconds() == 0 {
		errs = append(errs, fmt.Errorf("'refresh' must be >0 for generated resource: %v", generatedResource))
	}

	//Validate that all the action References are valid
	references, err := getReferencesFromAction(generatedResource.Action)
	if err != nil {
		errs = append(errs, err)
	}
	for _, r := range references {
		if !validateEntityReference(parsedTargetResource, r.entityType) {
			errs = append(errs, fmt.Errorf("invalid reference in action: %s", generatedResource.Action))
		}
	}

	return errs
}

type reference struct {
	entity     string
	entityType string
}

func getReferencesFromAction(action string) ([]*reference, error) {
	references := make([]*reference, 0)

	// Check if there is a reference to $resource in the action
	isMatch, err := regexp.MatchString(fmt.Sprintf(`\%s`, patterns.ResourceKW), action)
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
	re := regexp.MustCompile(fmt.Sprintf(`\%s\.(api|version|spec|artifact)($|/| )`, patterns.ResourceKW))
	match := re.FindAllString(action, -1)
	if len(match) == 0 {
		return nil, fmt.Errorf("invalid action: %s missing or incorrect entity in the reference", action)
	}

	// Construct a list of entity: entityType for all the references
	for _, m := range match {
		// entity = $resource.api, extract the entityType as "api"
		entity := strings.TrimRight(m, " /")
		entityType := entity[len(patterns.ResourceKW)+1:]
		references = append(references, &reference{entity: entity, entityType: entityType})
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

	resource, err := patterns.ParseResourcePattern(resourceName)
	if err != nil {
		return "", fmt.Errorf("error generating command, invalid resourceName: %s", resourceName)
	}

	for _, r := range references {
		entityVal := ""
		switch r.entityType {
		case "api":
			entityVal = resource.Api()
		case "version":
			entityVal = resource.Version()
		case "spec":
			entityVal = resource.Spec()
		case "artifact":
			entityVal = resource.Artifact()
		}

		if len(entityVal) == 0 {
			return "", fmt.Errorf("error generating command, cannot derive args for action. Invalid action: %s", action)
		}
		action = strings.ReplaceAll(action, r.entity, entityVal)
	}

	return action, nil
}
