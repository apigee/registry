package tools

import (
	"strings"
)

// ParentAndIdOfResourceNamed returns the name of a resource's parent and its resource ID.
func ParentAndIdOfResourceNamed(name string) (string, string) {
	parts := strings.Split(name, "/")
	last := len(parts) - 1
	return strings.Join(parts[0:last-1], "/"), parts[last]
}
