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

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
)

type Action struct {
	Command           string
	GeneratedResource string
	RequiresReceipt   bool
}

func ProcessManifest(
	ctx context.Context,
	client connection.Client,
	projectID string,
	manifest *rpc.Manifest) []*Action {

	var actions []*Action
	for _, resource := range manifest.GeneratedResources {
		log.Debugf(ctx, "Processing entry: %v", resource)

		err := ValidateResourceEntry(resource)
		if err != nil {
			log.FromContext(ctx).WithError(err).Debugf("Skipping resource: %q invalid resource pattern", resource)
			continue
		}

		newActions, err := processManifestResource(ctx, client, projectID, resource)
		if err != nil {
			log.FromContext(ctx).WithError(err).Debugf("Skipping resource: %q", resource)
			continue
		}
		actions = append(actions, newActions...)
	}

	return actions
}

func processManifestResource(
	ctx context.Context,
	client connection.Client,
	projectID string,
	generatedResource *rpc.GeneratedResource) ([]*Action, error) {
	// Generate dependency map
	resourcePattern := fmt.Sprintf("projects/%s/locations/global/%s", projectID, generatedResource.Pattern)
	dependencyMaps := make([]map[string]time.Time, 0, len(generatedResource.Dependencies))
	for _, dependency := range generatedResource.Dependencies {
		dMap, err := generateDependencyMap(ctx, client, resourcePattern, dependency, projectID)
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
	client connection.Client,
	resourcePattern string,
	dependency *rpc.Dependency,
	projectID string) (map[string]time.Time, error) {
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

	// Extend the dependency pattern if it contains $resource.api like pattern
	extDependencyQuery, err := extendDependencyPattern(resourcePattern, dependency.Pattern, projectID)
	if err != nil {
		return nil, err
	}

	// Fetch resources using the extDependencyQuery
	sourceList, err := ListResources(ctx, client, extDependencyQuery, dependency.Filter)
	if err != nil {
		return nil, err
	}

	for _, source := range sourceList {
		group, err := getEntityKey(dependency.Pattern, source.GetResourceName())
		if err != nil {
			return nil, err
		}

		sourceTime := source.GetUpdateTimestamp()
		maxUpdateTime, exists := sourceMap[group]
		if !exists || maxUpdateTime.Before(sourceTime) {
			sourceMap[group] = sourceTime
		}
	}

	if len(sourceMap) == 0 {
		return nil, fmt.Errorf("no resources found for pattern: %s, filer: %s", extDependencyQuery, dependency.Filter)
	}

	return sourceMap, nil

}

func generateActions(
	ctx context.Context,
	client connection.Client,
	resourcePattern string,
	filter string,
	dependencyMaps []map[string]time.Time,
	generatedResource *rpc.GeneratedResource) []*Action {
	actions := make([]*Action, 0)

	// Calculate actions only if dependencies are non-empty
	if len(dependencyMaps) > 0 {
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
	}

	return actions

}

// Go over the list of existing target resources to figure out which ones need an update.
func generateUpdateActions(
	ctx context.Context,
	client connection.Client,
	resourcePattern string,
	filter string,
	dependencyMaps []map[string]time.Time,
	generatedResource *rpc.GeneratedResource) ([]*Action, map[string]bool, error) {

	// Visited tracks the parents of target resources which were already generated.
	visited := make(map[string]bool)
	actions := make([]*Action, 0)

	// Generate resource list
	resourceList, err := ListResources(ctx, client, resourcePattern, filter)
	if err != nil {
		return nil, nil, err
	}

	// Iterate over a list of existing target resources to generate update actions
	for _, targetResource := range resourceList {
		visited[targetResource.GetResourceName().GetParent()] = true

		takeAction, err := needsUpdate(
			targetResource.GetResourceName(),
			targetResource.GetUpdateTimestamp(),
			dependencyMaps,
			generatedResource,
			false,
		)

		if err != nil {
			log.Errorf(ctx, "%s", err)
			continue
		}

		if takeAction {
			cmd, err := generateCommand(generatedResource.Action, targetResource.GetResourceName().String())
			if err != nil {
				return nil, nil, fmt.Errorf("Cannot generate command: %s", err)
			}
			a := &Action{
				Command:           cmd,
				GeneratedResource: targetResource.GetResourceName().String(),
				RequiresReceipt:   generatedResource.Receipt,
			}
			actions = append(actions, a)
		}

	}

	return actions, visited, nil
}

// For the target resources which do not exist in the registry yet,
//we will use the parent resources to derive which new target resources should be created.
func generateCreateActions(
	ctx context.Context,
	client connection.Client,
	resourcePattern string,
	dependencyMaps []map[string]time.Time,
	generatedResource *rpc.GeneratedResource,
	visited map[string]bool) ([]*Action, error) {

	parsedResourcePattern, err := parseResourcePattern(resourcePattern)
	if err != nil {
		return nil, err
	}
	parentName := parsedResourcePattern.GetParent()

	// If parent is a project, we can't list projects since this is registry client command.
	// Since the manifest definition is scoped  only for a particular project,
	// there will be only one target resource in this case.
	// There are two cases where this might happen:
	// 1. Target resource is a project level artifact "projects/demo/locations/global/artifacts/serach-index"
	//    extracted parent will be "projects/demo/locations/global"
	// 1. Target resource is an api "projects/demo/locations/global/apis/petstore"
	//    extracted parent will be "projects/demo/locations/global"
	if strings.HasSuffix(parentName, "locations/global") {
		// Return if this parent was already visited.
		if visited[parentName] {
			return nil, nil
		}

		// Since the GeneratedResource is non-existent here,
		// we will have to derive the exact name of the target resource.
		targetResourceName, err := resourceNameFromParent(resourcePattern, parentName)
		if err != nil {
			return nil, fmt.Errorf("Cannot generate target resourceName to be created. Error: %s", err)
		}

		takeAction, err := needsCreate(
			targetResourceName,
			dependencyMaps,
			generatedResource,
		)

		if err != nil {
			return nil, err
		} else if !takeAction {
			return nil, nil
		}

		cmd, err := generateCommand(generatedResource.Action, targetResourceName.String())
		if err != nil {
			return nil, fmt.Errorf("Cannot generate command: %s", err)
		}

		return []*Action{
			{
				Command:           cmd,
				GeneratedResource: targetResourceName.String(),
				RequiresReceipt:   generatedResource.Receipt,
			},
		}, nil
	}

	// If parent resource is not a project, then go through all the non-visited parents.
	// We don't pass the filter here because the filter is for the target resource and not it's parent.
	parentList, err := ListResources(ctx, client, parentName, "")
	if err != nil {
		return nil, err
	}

	actions := make([]*Action, 0)

	for _, parent := range parentList {
		// Skip if this parent was already visited.
		if visited[parent.GetResourceName().String()] {
			continue
		}
		// Since the GeneratedResource is non-existent here,
		// we will have to derive the exact name of the target resource
		targetResourceName, err := resourceNameFromParent(resourcePattern, parent.GetResourceName().String())
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
			return nil, fmt.Errorf("Cannot generate command: %s", err)
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
	targetResourceName ResourceName,
	targetResourceTime time.Time,
	dependencyMaps []map[string]time.Time,
	generatedResource *rpc.GeneratedResource,
	createMode bool) (bool, error) {
	for i, dependency := range generatedResource.Dependencies {
		dMap := dependencyMaps[i]
		// Get the entity to look for in dependencyMap
		entityKey, err := getEntityKey(dependency.Pattern, targetResourceName)
		if err != nil {
			// This means that there is error in the pattern definition, hence return
			return false, fmt.Errorf("cannot match resource with dependency. Error: %s", err.Error())
		}

		// All the dependencies should be present to generate an action.
		maxUpdateTime, ok := dMap[entityKey]
		if !ok {
			return false, nil
		}

		if maxUpdateTime.After(targetResourceTime) {
			return true, nil // Take action if atleast one dependency timestamp is later than resource timestamp
		}
	}
	return false, nil
}

func needsCreate(
	targetResourceName ResourceName,
	dependencyMaps []map[string]time.Time,
	generatedResource *rpc.GeneratedResource) (bool, error) {
	for i, dependency := range generatedResource.Dependencies {
		dMap := dependencyMaps[i]
		// Get the entity to look for in dependencyMap
		entityKey, err := getEntityKey(dependency.Pattern, targetResourceName)
		if err != nil {
			// This means that there is error in the pattern definition, hence return
			return false, fmt.Errorf("cannot match resource with dependency. Error: %s", err.Error())
		}

		// All the dependencies should be present to generate an action.
		if _, ok := dMap[entityKey]; !ok {
			return false, nil
		}
	}
	return true, nil
}
