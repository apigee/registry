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
	resource *rpc.GeneratedResource) ([]*Action, error) {
	// Generate dependency map
	resourcePattern := fmt.Sprintf("projects/%s/locations/global/%s", projectID, resource.Pattern)
	dependencyMaps := make([]map[string]time.Time, 0, len(resource.Dependencies))
	for _, dependency := range resource.Dependencies {
		dMap, err := generateDependencyMap(ctx, client, resourcePattern, dependency, projectID)
		if err != nil {
			return nil, fmt.Errorf("error while generating dependency map for %v: %s", dependency, err)
		}
		dependencyMaps = append(dependencyMaps, dMap)
	}

	// Generate resource list
	resourceList, err := ListResources(ctx, client, resourcePattern, resource.Filter)
	if err != nil {
		return nil, err
	}

	// Generate actions to update target resources
	actions, err := generateActions(
		ctx, client, projectID, resourcePattern, resourceList, dependencyMaps, resource)
	if err != nil {
		return nil, err
	}

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
		group, err := getEntityKey(dependency.Pattern, source)
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
	projectID string,
	resourcePattern string,
	resourceList []ResourceInstance,
	dependencyMaps []map[string]time.Time,
	generatedResource *rpc.GeneratedResource) ([]*Action, error) {

	actions := make([]*Action, 0)

	// Calculate actions only if dependencies are non-empty
	if len(dependencyMaps) > 0 {
		updateActions, visited, err := generateUpdateActions(ctx, client, resourcePattern, resourceList, dependencyMaps, generatedResource)
		// generateUpdateActions will return errors only if it finds errors in pattern definitions.
		// Hence return and avoid generating createAction with incorrect patterns.
		if err != nil {
			return nil, err
		}
		actions = append(actions, updateActions...)

		createActions, err := generateCreateActions(ctx, client, projectID, resourcePattern, resourceList, dependencyMaps, generatedResource, visited)
		if err != nil {
			// Return the updateActions generated previously. Do not report failure since this is a partial success.
			return actions, nil
		}
		actions = append(actions, createActions...)
	}

	return actions, nil

}

func generateUpdateActions(
	ctx context.Context,
	client connection.Client,
	resourcePattern string,
	resourceList []ResourceInstance,
	dependencyMaps []map[string]time.Time,
	generatedResource *rpc.GeneratedResource) ([]*Action, map[string]bool, error) {

	visited := make(map[string]bool)
	actions := make([]*Action, 0)

	// Iterate over a list of existing target resources to generate update actions
	for _, resource := range resourceList {
		resourceTime := resource.GetUpdateTimestamp()

		takeAction := false

		// Evaluate this resource against each dependency source pattern
		for i, dependency := range generatedResource.Dependencies {
			dMap := dependencyMaps[i]
			// Get the entity to look for in dependencyMap
			entity, err := getEntityKey(dependency.Pattern, resource)
			if err != nil {
				// This means that there is error in the pattern definition, hence return
				return nil, nil, fmt.Errorf("cannot match resource with dependency. Error: %s", err.Error())
			}

			if maxUpdateTime, ok := dMap[entity]; ok {
				// Take action if dependency timestamp is later than resource timestamp
				if maxUpdateTime.After(resourceTime) {
					takeAction = true
				}
				visited[entity] = true
			} else {
				// For a given resource, each of it's defined dependency entity should be present.
				// If any one of the dependency entities is missing, avoid calculating any action for the resource
				takeAction = false
				break
			}
		}

		if takeAction {
			cmd, err := generateCommand(generatedResource.Action, resource.GetName())
			if err != nil {
				// This means that there is error in action definition, hence return
				return nil, nil, err
			}
			action := &Action{
				Command:           cmd,
				GeneratedResource: resource.GetName(),
				RequiresReceipt:   generatedResource.Receipt,
			}
			actions = append(actions, action)
		}
	}

	return actions, visited, nil
}

// Check patterns where resources do not exist in the registry. Here new resources will be generated
// for the dependencies which were not visited in the update actions flow.
func generateCreateActions(
	ctx context.Context,
	client connection.Client,
	projectID string,
	resourcePattern string,
	resourceList []ResourceInstance,
	dependencyMaps []map[string]time.Time,
	generatedResource *rpc.GeneratedResource,
	visited map[string]bool) ([]*Action, error) {

	actions := make([]*Action, 0)

	// Find the index of dependency which has reference to the lowest entity in $resource.
	entityType, index, err := minRefIndexFromDependency(generatedResource)
	if err != nil {
		return nil, fmt.Errorf("Cannot generate actions for non-existent target resources, error: %s", err)
	}

	// If the lowest reference entityType is empty, that means there is no $resource reference present in the dependencies.
	// Since lowest reference entity for action and dependencies have to match, action string will have no references to $resource.
	// Since there are no references, there is nothing to substitue and hance we can return the action string as is.
	if entityType == "" {
		if _, ok := visited[""]; !ok {
			// Verify that there are no $resource references in the action string
			references, err := getReferencesFromAction(generatedResource.Action)
			if err != nil {
				return nil, fmt.Errorf("Cannot generate actions invalid action string: %s", err)
			}
			if len(references) == 0 {
				return append(actions, &Action{
					Command:           generatedResource.Action,
					GeneratedResource: resourcePattern,
					RequiresReceipt:   generatedResource.Receipt,
				}), nil
			} else {
				return nil, fmt.Errorf("Invalid GeneratedResource definition %v", generatedResource)
			}
		}
	} else { //When dependencies and action have references to $resource entities

		// Iterate over all the non visited lowestEntities in this map
		for lowestEntityKey := range dependencyMaps[index] {
			// Get the resource with lowest entity reference and use that to derive other dependencies and generate action.
			//ListResources will return only one resource
			list, err := ListResources(ctx, client, lowestEntityKey, "")
			if err != nil || len(list) == 0 {
				log.Debugf(ctx, "Error while generating createAction: Resource %s cannot be fetched", lowestEntityKey)
				continue
			}
			lowestEntityResource := list[0]

			takeAction := true
			if _, ok := visited[lowestEntityKey]; !ok {
				for i, dMap := range dependencyMaps {
					// in a particular group resources for each dependencyPattern must exist
					// for the controller to generate new targeted resource.
					// Example:
					// generated_resource: apis/-/artifacts/summary
					// - dependencies:
					//   - pattern: $resource.api/artifacts/complexity
					//   - pattern: $resource.api/artifacts/vocabulary
					// for a particular api group, "petstore"
					// both apis/petstore/artifacts/complexity and apis/petstore/artifacts/vocabulary should be present
					// Get the entity to look for in dependencyMap
					entity, err := getEntityKey(generatedResource.Dependencies[i].Pattern, lowestEntityResource)
					if err != nil {
						takeAction = false
						break
					}
					_, dependencyExists := dMap[entity]
					if !dependencyExists {
						takeAction = false
						break
					}
				}
			} else {
				takeAction = false
			}

			if takeAction {
				// Since the GeneratedResource is non-existent here,
				// we will have to derive the exact name of the target resource
				resourceName, err := resourceNameFromEntityKey(
					resourcePattern, lowestEntityKey)
				if err != nil {
					log.Debugf(ctx, "skipping entry for %q, cannot derive target resource name for pattern: %q", lowestEntityKey, resourcePattern)
					continue
				}
				cmd, err := generateCommand(generatedResource.Action, resourceName)
				if err != nil {
					log.Debugf(ctx, "Error generating action for resource: %q", resourceName)
					continue
				}

				action := &Action{
					Command:           cmd,
					GeneratedResource: resourceName,
					RequiresReceipt:   generatedResource.Receipt,
				}
				actions = append(actions, action)
			}

		}
	}

	return actions, nil
}
