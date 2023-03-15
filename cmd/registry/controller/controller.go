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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/pkg/application/controller"
	"github.com/apigee/registry/pkg/log"
)

type Action struct {
	Command           string
	GeneratedResource string
	RequiresReceipt   bool
}

func ProcessManifest(
	ctx context.Context,
	client listingClient,
	projectID string,
	manifest *controller.Manifest,
	maxActions int) []*Action {
	var actions []*Action
	//Check for errors in manifest
	errs := ValidateManifest(fmt.Sprintf("projects/%s/locations/global", projectID), manifest)
	if len(errs) > 0 {
		for _, err := range errs {
			log.FromContext(ctx).WithError(err).Debugf("Error in manifest")
		}
	}

	for _, resource := range manifest.GeneratedResources {
		log.Debugf(ctx, "Processing entry: %v", resource)

		errs := validateGeneratedResourceEntry(fmt.Sprintf("projects/%s/locations/global", projectID), resource)
		if len(errs) > 0 {
			log.FromContext(ctx).Debugf("Skipping resource: %q", resource)
			continue
		}

		newActions, err := processManifestResource(ctx, client, projectID, resource)
		if err != nil {
			log.FromContext(ctx).WithError(err).Debugf("Skipping resource: %q", resource)
			continue
		}
		actions = append(actions, newActions...)

		if len(actions) >= maxActions {
			log.FromContext(ctx).Debugf("Reached max actions limit %d", maxActions)
			break
		}
	}

	maxLength := len(actions)
	if maxLength > maxActions {
		maxLength = maxActions
	}

	return actions[:maxLength]
}

func processManifestResource(
	ctx context.Context,
	client listingClient,
	projectID string,
	generatedResource *controller.GeneratedResource) ([]*Action, error) {
	resourcePattern := fmt.Sprintf("projects/%s/locations/global/%s", projectID, generatedResource.Pattern)
	// Generate dependency map
	dependencyMaps := make([]map[string]time.Time, 0, len(generatedResource.Dependencies))
	for _, dependency := range generatedResource.Dependencies {
		dMap, err := generateDependencyMap(ctx, client, resourcePattern, dependency)
		if err != nil {
			return nil, fmt.Errorf("error while generating dependency map for %v: %s", dependency, err)
		}
		dependencyMaps = append(dependencyMaps, dMap)
	}

	// Generate actions to create and update target resources
	actions := generateActions(
		ctx, client, resourcePattern, generatedResource.Filter, dependencyMaps, generatedResource)

	return actions, nil
}

func generateDependencyMap(
	ctx context.Context,
	client listingClient,
	resourcePattern string,
	dependency *controller.Dependency) (map[string]time.Time, error) {
	// Creates a map of the resources to group them into corresponding buckets
	// of match pattern which store the maxTimestamp
	// An example entry will look like this:
	// dependencyPattern: $resource.api/versions/-/specs/-   ($resource.api is the match)
	// Map:
	// - key: projects/demo/locations/global/apis/petstore
	//   value: maxUpdateTime: 00:00:00
	// - key: projects/demo/locations/global/apis/wordnik.com
	//   value: maxUpdateTime: 00:00:00

	sourceMap := make(map[string]time.Time)

	resourceName, err := patterns.ParseResourcePattern(resourcePattern)
	if err != nil {
		return nil, err
	}

	// Extend the dependency pattern if it contains $resource.api like pattern
	extDependencyName, err := patterns.SubstituteReferenceEntity(dependency.Pattern, resourceName)
	if err != nil {
		return nil, err
	}

	// Fetch resources using the extDependencyQuery
	sourceList, err := listResources(ctx, client, extDependencyName.String(), dependency.Filter)
	if err != nil {
		return nil, err
	}

	for _, source := range sourceList {
		group, err := patterns.GetReferenceEntityValue(dependency.Pattern, source.ResourceName())
		if err != nil {
			return nil, err
		}

		sourceTime := source.UpdateTimestamp()
		maxUpdateTime, exists := sourceMap[group]
		if !exists || maxUpdateTime.Before(sourceTime) {
			sourceMap[group] = sourceTime
		}
	}

	if len(sourceMap) == 0 {
		return nil, fmt.Errorf("no resources found for pattern: %s, filer: %s", extDependencyName.String(), dependency.Filter)
	}

	return sourceMap, nil
}

func generateActions(
	ctx context.Context,
	client listingClient,
	resourcePattern string,
	filter string,
	dependencyMaps []map[string]time.Time,
	generatedResource *controller.GeneratedResource) []*Action {
	actions := make([]*Action, 0)

	updateActions, visited, err := generateUpdateActions(ctx, client, resourcePattern, filter, dependencyMaps, generatedResource)
	if err != nil {
		log.Errorf(ctx, "Error while generating UpdateActions: %s", err)
	}
	actions = append(actions, updateActions...)

	createActions, err := generateCreateActions(ctx, client, resourcePattern, dependencyMaps, generatedResource, visited)
	if err != nil {
		log.Errorf(ctx, "Error while generating CreateActions: %s", err)
	}
	actions = append(actions, createActions...)

	return actions
}

// Go over the list of existing target resources to figure out which ones need an update.
func generateUpdateActions(
	ctx context.Context,
	client listingClient,
	resourcePattern string,
	filter string,
	dependencyMaps []map[string]time.Time,
	generatedResource *controller.GeneratedResource) ([]*Action, map[string]bool, error) {
	// Visited tracks the parents of target resources which were already generated.
	visited := make(map[string]bool)
	actions := make([]*Action, 0)

	// Generate resource list
	resourceList, err := listResources(ctx, client, resourcePattern, filter)
	if err != nil {
		return nil, nil, err
	}

	// Iterate over a list of existing target resources to generate update actions
	for _, targetResource := range resourceList {
		visited[targetResource.ResourceName().ParentName().String()] = true

		takeAction, err := needsUpdate(
			targetResource.ResourceName(),
			targetResource.UpdateTimestamp(),
			dependencyMaps,
			generatedResource,
		)

		if err != nil {
			log.Errorf(ctx, "%s", err)
			continue
		}

		if takeAction {
			cmd, err := generateCommand(generatedResource.Action, targetResource.ResourceName().String())
			if err != nil {
				return nil, nil, fmt.Errorf("cannot generate command: %s", err)
			}
			a := &Action{
				Command:           cmd,
				GeneratedResource: targetResource.ResourceName().String(),
				RequiresReceipt:   generatedResource.Receipt,
			}
			actions = append(actions, a)
		}
	}

	return actions, visited, nil
}

// Constructs a CEL filter to exclude resources with visited parents.
// Makes use of `e.all(x,p)` macro as defined here: https://github.com/google/cel-spec/blob/master/doc/langdef.md#macros
// The filter excludes resources whose `name` property is equal to any of the visited parent names.
//
// For example, consider a visited map of parents which are apis:
//
//	{
//	    "projects/demo/locations/global/apis/example-api1": true,
//	    "projects/demo/locations/global/apis/example-api2": true,
//	}
//
// The resulting CEL filter will be:
// ["projects/demo/locations/global/apis/example-api1","projects/demo/locations/global/apis/example-api2"].all(parent, !(name==parent))
//
// Note: The `bool` values in the input map are ignored. The filter will use every map key.
func excludeVisitedParents(v map[string]bool) string {
	// Wrap each string with quotes and join them with commas to build a JSON string array.
	jsonStrings := make([]string, 0, len(v))
	for parent := range v {
		// filter ignores revisions
		jsonStrings = append(jsonStrings, fmt.Sprintf("%q", strings.Split(parent, "@")[0]))
	}
	return fmt.Sprintf("[%s].all(parent, !(name==parent))", strings.Join(jsonStrings, ","))
}

// For the target resources which do not exist in the registry yet,
// we will use the parent resources to derive which new target resources should be created.
func generateCreateActions(
	ctx context.Context,
	client listingClient,
	resourcePattern string,
	dependencyMaps []map[string]time.Time,
	generatedResource *controller.GeneratedResource,
	visited map[string]bool) ([]*Action, error) {
	var parentList []patterns.ResourceInstance

	parsedResourcePattern, err := patterns.ParseResourcePattern(resourcePattern)
	if err != nil {
		return nil, err
	}

	parentName := parsedResourcePattern.ParentName()
	switch parentName.(type) {
	case patterns.ProjectName:
		// If parent is a project, we can't list projects since this is registry client command.
		// Since the manifest definition is scoped  only for a particular project,
		// there will be only one target resource in this case.
		// There are two cases where this might happen:
		// 1. Target resource is a project level artifact "projects/demo/locations/global/artifact/search-index"
		//    extracted parent will be "projects/demo/locations/global"
		// 1. Target resource is an api "projects/demo/locations/global/apis/petstore"
		//    extracted parent will be "projects/demo/locations/global"
		// Return if this parent was already visited.
		if visited[parentName.String()] {
			return nil, nil
		}
		parentList = []patterns.ResourceInstance{
			patterns.ProjectResource{
				ProjectName: parentName.String(),
			},
		}

	default:
		filter := excludeVisitedParents(visited)
		parentList, err = listResources(ctx, client, parentName.String(), filter)
		if err != nil {
			return nil, err
		}
	}

	actions := make([]*Action, 0)
	for _, parent := range parentList {
		// Since the GeneratedResource is nonexistent here,
		// we will have to derive the exact name of the target resource
		targetResourceName, err := patterns.FullResourceNameFromParent(resourcePattern, parent.ResourceName().String())
		if err != nil {
			return nil, err
		}

		takeAction, err := needsCreate(
			targetResourceName,
			dependencyMaps,
			generatedResource,
		)

		if err != nil {
			return nil, err
		} else if !takeAction {
			continue
		}

		cmd, err := generateCommand(generatedResource.Action, targetResourceName.String())
		if err != nil {
			return nil, fmt.Errorf("cannot generate command: %s", err)
		}
		a := &Action{
			Command:           cmd,
			GeneratedResource: targetResourceName.String(),
			RequiresReceipt:   generatedResource.Receipt,
		}
		actions = append(actions, a)
	}

	return actions, nil
}

func needsUpdate(
	targetResourceName patterns.ResourceName,
	targetResourceTime time.Time,
	dependencyMaps []map[string]time.Time,
	generatedResource *controller.GeneratedResource) (bool, error) {
	// Check "refresh" first to decide whether to take action or not.
	if generatedResource.Refresh != nil && targetResourceTime.Add(generatedResource.Refresh.AsDuration()).Before(time.Now()) {
		return true, nil
	}
	// Check for dependencies otherwise
	for i, dependency := range generatedResource.Dependencies {
		dMap := dependencyMaps[i]
		// Get the entity to look for in dependencyMap
		entityKey, err := patterns.GetReferenceEntityValue(dependency.Pattern, targetResourceName)
		if err != nil {
			// This means that there is error in the pattern definition, hence return
			return false, fmt.Errorf("cannot match resource with dependency. Error: %s", err.Error())
		}

		// All the dependencies should be present to generate an action.
		maxUpdateTime, ok := dMap[entityKey]
		if !ok {
			return false, nil
		}

		// Take action if the target resource is less than n seconds newer compared to the dependencies, where n=thresholdSeconds.
		// https://github.com/apigee/registry/issues/641
		if maxUpdateTime.Add(patterns.ResourceUpdateThreshold).After(targetResourceTime) {
			return true, nil
		}
	}
	return false, nil
}

func needsCreate(
	targetResourceName patterns.ResourceName,
	dependencyMaps []map[string]time.Time,
	generatedResource *controller.GeneratedResource) (bool, error) {
	// Take action if "refresh" is set and > 0
	if generatedResource.Refresh != nil && generatedResource.Refresh.AsDuration().Seconds() > 0 {
		return true, nil
	}
	// Check for dependencies otherwise
	for i, dependency := range generatedResource.Dependencies {
		dMap := dependencyMaps[i]
		// Get the entity to look for in dependencyMap
		entityVal, err := patterns.GetReferenceEntityValue(dependency.Pattern, targetResourceName)
		if err != nil {
			// This means that there is error in the pattern definition, hence return
			return false, fmt.Errorf("cannot match resource with dependency. Error: %s", err.Error())
		}

		// All the dependencies should be present to generate an action.
		if _, ok := dMap[entityVal]; !ok {
			return false, nil
		}
	}
	return true, nil
}
