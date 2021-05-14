package  controller

import (
    "github.com/apigee/registry/cmd/control_loop/list"
    "github.com/apigee/registry/cmd/control_loop/resources"
    "reflect"
    "context"
	"github.com/apigee/registry/connection"
	"time"
	"fmt"
)

type ResourceCollection struct {
	maxUpdateTime time.Time
	resourceList *[]resources.Resource
}

func ProcessManifest(manifest *Manifest) ([]string, error) {

	var actions []string
	for _, entry := range manifest.Entries {
    	ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			return nil, err
		}		
		resourcePattern := fmt.Sprintf("projects/%s/%s", manifest.Project, entry.Resource)

		dependencyMaps := make([]map[string]*ResourceCollection, 0, 0)
		err = populateDependencyMaps(ctx, client, resourcePattern, entry.Dependencies, &dependencyMaps)
		if err != nil {
			return nil, err
		}

		cmds, err := updateResources(ctx, client, resourcePattern, entry.Dependencies, dependencyMaps, entry.Action)
		if err != nil {
			return nil, err
		}
		if len(cmds) > 0 {
			actions = append(actions, cmds...)
		}
	}

	return actions, nil
}

func populateDependencyMaps(
	ctx context.Context,
	client connection.Client,
	resourcePattern string,
	dependencies []Dependency,
	dependencyMaps *[]map[string]*ResourceCollection) error {
	for _, d := range dependencies {
		sourceMap := make(map[string]*ResourceCollection)
		// Extend the source pattern if it contains $resource.api like pattern
		sourcePattern, err := ParseSourcePattern(resourcePattern, d.Source)
		if err != nil {
			return err
		}

		// Fetch resources using the sourcePattern
		sourceList, err :=  list.ListResources( ctx, client, sourcePattern, d.Filter)
		if err != nil {
			return err
		}

		// Extract the attribute which will be used
		// to group the contents of the sourceList into a map
		groupFuncName := ParseGroupFunc(d.Source)
		
		// Build source map
		for _, source := range sourceList {
			// Group all the resources in one group by default
			groupBy := ""

			if len(groupFuncName) > 0 {
				groupFunc := reflect.ValueOf(source).MethodByName(groupFuncName)
				groupBy = groupFunc.Call([]reflect.Value{})[0].String()
			}

			// Update the map with timestamp
			source_ts := source.GetUpdateTimestamp()
			if collection, ok := sourceMap[groupBy]; !ok {
				sourceMap[groupBy] = &ResourceCollection {
					maxUpdateTime: source_ts,
					resourceList: &[]resources.Resource{},
				}
			} else if (*collection).maxUpdateTime.Before(source_ts) {
					(*collection).maxUpdateTime = source_ts
			}

			temp := (*sourceMap[groupBy]).resourceList
			(*temp) = append((*temp), source)
		}
		(*dependencyMaps) = append((*dependencyMaps), sourceMap)

	}

	return nil
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
		r_ts := resource.GetUpdateTimestamp()
		takeAction := false
		var args []string

		// Evaluate this resource against each dependency source pattern
		for i, dependency := range dependencies {
			dMap := dependencyMaps[i]
			// Get the group to look for in dependencyMap
			groupFuncName := ParseGroupFunc(dependency.Source)
			groupBy := ""

			if len(groupFuncName) > 0 {
				groupFunc := reflect.ValueOf(resource).MethodByName(groupFuncName)
				groupBy = groupFunc.Call([]reflect.Value{})[0].String()
			}

			if collection, ok := dMap[groupBy]; ok {
			// Take action if dependency timestamp is later than resource timestamp
				if (*collection).maxUpdateTime.After(r_ts) {
					takeAction = true
				}
				visited[groupBy] = true
				argVal := DeriveArgsFromResources(i, (*(*collection).resourceList)[0], action)
				args = append(args, argVal)
			} else {
				takeAction = false
				break
			}
		}

		if takeAction {
			GenerateCommand(action, args, &cmds)
		}
	}

	// Iterate over first dependency source and evaluate that against remaining dependencies
	if len(dependencyMaps) > 0 {
		dMap0 := dependencyMaps[0]
		for key := range dMap0 {
			takeAction := true
			var args []string
			if _, ok := visited[key]; !ok {
				for i, dMap := range dependencyMaps {
					collection, ok := dMap[key]
					if ok {
						argVal := DeriveArgsFromResources(i, (*(*collection).resourceList)[0], action)
						args = append(args, argVal)
					} else {
						takeAction = false
						break
					}
				}
			} else {
				takeAction= false
			}

			if takeAction {
				GenerateCommand(action, args, &cmds)
			}

		}
	}

	return cmds, nil
}