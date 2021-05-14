package controller

import (
	"regexp"
	"strings"
	"errors"
	"fmt"
	"github.com/apigee/registry/cmd/control_loop/resources"
    "reflect"
)

func ParseSourcePattern(
	rPattern string,
	sPattern string) (string, error) {
	
	rEntity := extractEntity(sPattern)

	//If there was an attribute present, replace it with the pattern from resource pattern 
	if len(rEntity) > 0{
		valEntity, err := extractEntityValFromPattern(rPattern, rEntity)
		if err != nil {
			return "", errors.New("Invalid pattern")
		}
		extendedSPattern := strings.Replace(
			sPattern,
			fmt.Sprintf("$resource.%s", rEntity),
			valEntity,
			1)
		return extendedSPattern, nil
	}

	// else return the original source pattern 
	return sPattern, nil

}

func extractEntity(pattern string) string {
	// Check for $resource.api kind of pattern in the "sourcePattern"
	// and retrieve the "api" attribute. This attribute will be used 
	// to fetch the corresponding pattern from the resource.
	entityRegex := `\$resource\.(api|version|spec|artifact)`
	re := regexp.MustCompile(entityRegex)
	entities := re.FindStringSubmatch(pattern)
	entity := ""

	if len(entities) > 1 {
		entity = entities[1]
	}

	return entity
}

func extractEntityValFromPattern(pattern string, entity string) (string, error) {
	re := regexp.MustCompile(fmt.Sprintf(`.*/%ss/[^/]*`, entity))
	entityVal := re.FindString(pattern)
	if len(entityVal) > 0 {
		return entityVal, nil
	}
	return "", errors.New("Empty entity")
}

func ParseGroupFunc(pattern string) string {
	entity := extractEntity(pattern)
	if len(entity) > 0 {
		return fmt.Sprintf("Get%s", strings.Title(entity))
	}

	return ""
}

func DeriveArgsFromResources(
	index int,
	resource resources.Resource,
	action string) string {
	entityRegex := fmt.Sprintf(`.*(\$source%d(\.api|\.version|\.spec|\.artifact)?)`, index)
	re := regexp.MustCompile(entityRegex)
	match := re.FindStringSubmatch(action)

	// The above func FindStringSubmatch will always return a slice of size 3
	// Example:
	// re.FindStringSubmatch("compute lint $source0") = ["compute lint $source0", "$source0", ""]
	// re.FindStringSubmatch("compute lint $source0.spec") = ["compute lint $source0.spec", "$source0", ".spec"]

	// if the source is present with an entity. Eg: $source0.api
	if len(match[2]) > 0 {
		entityFuncName := fmt.Sprintf("Get%s", strings.Title(match[2][1:]))
		entityFunc := reflect.ValueOf(resource).MethodByName(entityFuncName)
		entityVal := entityFunc.Call([]reflect.Value{})[0].String()
		return entityVal
	} else if len(match[1]) > 0 { //if only source is present. Eg: $source0
		return resource.GetName()
	}

	return ""
}

func GenerateCommand(action string, args []string, cmds *[]string) {
	for i, arg := range args {
		// Find the pattern to replace with argValue
		argRegex := regexp.MustCompile(fmt.Sprintf(`\$source%d(\.(api|version|spec|artifact))?`, i))
		argPattern := argRegex.FindString(action)
		action = strings.ReplaceAll(action, argPattern, arg) 
	}

	(*cmds) = append((*cmds), action)
}