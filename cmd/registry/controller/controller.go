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
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"log"
	"time"
)

type Action struct {
	Command           string
	GeneratedResource string
	Receipt           bool
}

type ResourceCollection struct {
	maxUpdateTime time.Time
	resourceList  []Resource
}

func ProcessManifest(
	ctx context.Context,
	client connection.Client,
	projectID string,
	manifest *rpc.Manifest) ([]*Action, error) {

	var actions []*Action
	for _, resource := range manifest.GeneratedResources {

		newActions, err := processManifestResource(ctx, client, projectID, resource)
		if err != nil {
			log.Printf("Skipping resource: %q\nGot error: %s", resource, err.Error())
		}
		actions = append(actions, newActions...)
	}

	return actions, nil
}

func processManifestResource(
	ctx context.Context,
	client connection.Client,
	projectID string,
	resource *rpc.GeneratedResource) ([]*Action, error) {
	// Generate dependency map
	resourcePattern := fmt.Sprintf("projects/%s/locations/global/%s", projectID, resource.Pattern)
	dependencyMaps := make([]map[string]ResourceCollection, 0, len(resource.Dependencies))
	for _, d := range resource.Dependencies {
		dMap, err := generateDependencyMap(ctx, client, resourcePattern, d.Pattern, d.Filter, projectID)
		if err != nil {
			log.Printf("Error while generating dependency map.\n Error: %s\n Skipping resource %+v", err.Error(), resource)
			continue
		}
		dependencyMaps = append(dependencyMaps, dMap)
	}

	// Generate resource list
	resourceList, err := ListResources(ctx, client, resourcePattern, resource.Filter)
	if err != nil {
		return nil, err
	}

	// Update target resources
	actions, err := generateActions(
		ctx, client, resourcePattern, resourceList, dependencyMaps, resource)
	if err != nil {
		return nil, err
	}

	return actions, nil
}

func generateDependencyMap(
	ctx context.Context,
	client connection.Client,
	resourcePattern,
	dependencyPattern,
	dependencyFilter string,
	projectID string) (map[string]ResourceCollection, error) {

	sourceMap := make(map[string]ResourceCollection)

	// Extend the source pattern if it contains $resource.api like pattern
	extDependencyPattern, err := ExtendSourcePattern(resourcePattern, dependencyPattern, projectID)
	if err != nil {
		return nil, err
	}

	// Fetch resources using the extDependencyPattern
	sourceList, err := ListResources(ctx, client, extDependencyPattern, dependencyFilter)
	if err != nil {
		return nil, err
	}

	for _, source := range sourceList {
		group, err := ExtractGroup(dependencyPattern, source)
		if err != nil {
			return nil, err
		}

		sourceTime := source.GetUpdateTimestamp()
		collection, exists := sourceMap[group]
		if !exists {
			collection = ResourceCollection{
				maxUpdateTime: sourceTime,
			}
		} else if collection.maxUpdateTime.Before(sourceTime) {
			collection.maxUpdateTime = sourceTime
		}

		collection.resourceList = append(collection.resourceList, source)
		sourceMap[group] = collection
	}

	return sourceMap, nil

}

func generateActions(
	ctx context.Context,
	client connection.Client,
	resourcePattern string,
	resourceList []Resource,
	dependencyMaps []map[string]ResourceCollection,
	generatedResource *rpc.GeneratedResource) ([]*Action, error) {

	visited := make(map[string]bool, 0)
	actions := make([]*Action, 0)

	for _, resource := range resourceList {
		resourceTime := resource.GetUpdateTimestamp()

		takeAction := false
		var args []Resource

		// Evaluate this resource against each dependency source pattern
		for i, dependency := range generatedResource.Dependencies {
			dMap := dependencyMaps[i]
			// Get the group to look for in dependencyMap
			group, err := ExtractGroup(dependency.Pattern, resource)
			if err != nil {
				return nil, fmt.Errorf("Cannot match resource with dependency. Error: %s", err.Error())
			}

			if collection, ok := dMap[group]; ok {
				// Take action if dependency timestamp is later than resource timestamp
				if collection.maxUpdateTime.After(resourceTime) {
					takeAction = true
				}
				visited[group] = true
				// TODO: Evaluate if append only the group or resource name should be enough
				args = append(args, collection.resourceList[0])
			} else {
				// For a given resource, each of it's defined dependency group should be present.
				// If any one of the dependency groups is missing, avoid calculating any action for the resource
				takeAction = false
				break
			}
		}

		if takeAction {
			cmd, err := GenerateCommand(generatedResource.Action, args)
			if err != nil {
				return nil, err
			}
			action := &Action{
				Command:           cmd,
				GeneratedResource: resource.GetName(),
				Receipt:           generatedResource.Receipt,
			}
			actions = append(actions, action)
		}
	}

	// Check patterns where resources do not exist in the registry. Here new resources will be generated
	// for the dependencies which were not visited in the above loop.
	// Iterate over first dependency source and evaluate that against remaining dependencies
	if len(dependencyMaps) > 0 {
		dMap0 := dependencyMaps[0]
		for key := range dMap0 {
			takeAction := true
			var args []Resource
			var resourceName string
			if _, ok := visited[key]; !ok {
				for _, dMap := range dependencyMaps {
					collection, ok := dMap[key]
					if ok {
						collectedResource := collection.resourceList[0]
						// Since the GeneratedResource is non-existent here,
						// we will have to derive the exact name of the target resource
						var err error
						resourceName, err = resourceNameFromDependency(
							resourcePattern, collectedResource)
						if err != nil {
							log.Printf("skipping entry for %q, cannot derive generated resource name from invalid pattern: %q", collectedResource, resourcePattern)
							takeAction = false
							break
						}
						args = append(args, collectedResource)
					} else {
						takeAction = false
						break
					}
				}
			} else {
				takeAction = false
			}

			if takeAction {
				cmd, err := GenerateCommand(generatedResource.Action, args)
				if err != nil {
					return nil, err
				}

				action := &Action{
					Command:           cmd,
					GeneratedResource: resourceName,
					Receipt:           generatedResource.Receipt,
				}
				actions = append(actions, action)
			}

		}
	}

	return actions, nil
}
