package  controller

import (
    "github.com/apigee/registry/cmd/control_loop/list"
    "github.com/apigee/registry/cmd/control_loop/resources"
    "context"
	"github.com/apigee/registry/connection"
	"time"
	"fmt"
	"log"
)

type ResourceCollection struct {
	maxUpdateTime time.Time
	resourceList *[]resources.Resource
}

func ProcessManifest(manifest *Manifest) ([]string, error) {

	var actions []string
	ctx := context.TODO()
	client, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}		
	for _, entry := range manifest.Entries {		
	
		newActions, err := processManifestEntry(ctx, client, manifest.Project, entry)
		if err != nil {
			log.Printf("Skipping entry: %q\nGot error: %s", entry, err.Error())

		}
		actions = append(actions, newActions...)
	}

	return actions, nil
}

func processManifestEntry(
	ctx context.Context,
	client connection.Client,
	project string,
	entry ManifestEntry) ([]string, error) {
	// Generate dependency map
	resourcePattern := fmt.Sprintf("projects/%s/%s", project, entry.Resource)
	dependencyMaps := make([]map[string]*ResourceCollection, 0, len(entry.Dependencies))
	for _, d := range entry.Dependencies {
		dMap, err := generateDependencyMap(ctx, client, resourcePattern, d.Source, d.Filter)
		if err != nil {
			log.Printf("Encountered error during updateResources().\n Error: %s\n Skipping entry %+v", err.Error(), entry)
			continue
		}
		dependencyMaps = append(dependencyMaps, dMap)
	}

	// Update target resources
	cmds, err := updateResources(ctx, client, resourcePattern, entry.Dependencies, dependencyMaps, entry.Action)
	if err != nil {
		return nil, err
	}

	return cmds, nil
}

func generateDependencyMap(
	ctx context.Context,
	client connection.Client,
	resourcePattern,
	sourcePattern,
	sourceFilter string) (map[string]*ResourceCollection, error) {

	sourceMap := make(map[string]*ResourceCollection)

	// Extend the source pattern if it contains $resource.api like pattern
	extSourcePattern, err := ExtendSourcePattern(resourcePattern, sourcePattern)
	if err != nil {
		return nil, err
	}

	// Fetch resources using the extSourcePattern
	sourceList, err :=  list.ListResources(ctx, client, extSourcePattern, sourceFilter)
	if err != nil {
		return nil, err
	}

	for _, source := range sourceList {
		group, err := ExtractGroup(sourcePattern, source)
		if err != nil {
			return nil, err
		}

		// Update the map with timestamp
		sourceTS := source.GetUpdateTimestamp()
		if collection, ok := sourceMap[group]; !ok {
			sourceMap[group] = &ResourceCollection {
				maxUpdateTime: sourceTS,
				resourceList: &[]resources.Resource{},
			}
		} else if (*collection).maxUpdateTime.Before(sourceTS) {
				(*collection).maxUpdateTime = sourceTS
		}

		temp := (*sourceMap[group]).resourceList
		(*temp) = append((*temp), source)
	}

	return sourceMap, nil

}

func updateResources(
	ctx context.Context,
	client connection.Client,
	resourcePattern string,
 	dependencies []Dependency,
	dependencyMaps []map[string]*ResourceCollection,
	action string) ([]string, error) {

	visited := make(map[string]bool, 0)
	cmds := make([]string, 0)


	resourceList, err := list.ListResources( ctx, client, resourcePattern, "")
	if err != nil {
		return nil, err
	}

	// TODO: Add more error handling in the two loops
	for _, resource := range resourceList {
		resourceTime := resource.GetUpdateTimestamp()

		takeAction := false
		var args []resources.Resource

		// Evaluate this resource against each dependency source pattern
		for i, dependency := range dependencies {
			dMap := dependencyMaps[i]
			// Get the group to look for in dependencyMap
			group, err := ExtractGroup(dependency.Source, resource)
			if err != nil {
				return nil, fmt.Errorf("Cannot match resource with source. Error: %s", err.Error())
			}

			if collection, ok := dMap[group]; ok {
			// Take action if dependency timestamp is later than resource timestamp
				if (*collection).maxUpdateTime.After(resourceTime) {
					takeAction = true
				}
				visited[group] = true
				// TODO: Evaluate if append only the group or resource name should be enough
				args = append(args, (*(*collection).resourceList)[0])
			} else {
				// For a given resource, each of it's defined dependency group should be present.
				// If any one of the dependency groups is missing, avoid calculating any action for the resource
				takeAction = false
				break
			}
		}

		if takeAction {
			cmd, err := GenerateCommand(action, args)
				if err != nil {
					return nil, err
				}
				cmds = append(cmds, cmd)
		}
	}

	// Iterate over first dependency source and evaluate that against remaining dependencies
	if len(dependencyMaps) > 0 {
		dMap0 := dependencyMaps[0]
		for key := range dMap0 {
			takeAction := true
			var args []resources.Resource
			if _, ok := visited[key]; !ok {
				for _, dMap := range dependencyMaps {
					collection, ok := dMap[key]
					if ok {
						args = append(args, (*(*collection).resourceList)[0])
					} else {
						takeAction = false
						break
					}
				}
			} else {
				takeAction = false
			}

			if takeAction {
				cmd, err := GenerateCommand(action, args)
				if err != nil {
					return nil, err
				}
				cmds = append(cmds, cmd)
			}

		}
	}

	return cmds, nil
}