// Copyright 2020 Google LLC. All Rights Reserved.

package names

import (
	"fmt"
	"regexp"
)

// SpecsRegexp returns a regular expression that matches a collection of specs.
func SpecsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex + "/products/" + NameRegex + "/versions/" + NameRegex + "/specs$")
}

// SpecRegexp returns a regular expression that matches a spec resource name.
func SpecRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex +
		"/products/" + NameRegex +
		"/versions/" + NameRegex +
		"/specs/" + NameRegex +
		RevisionRegex + "$")
}

// ParseParentVersion ...
func ParseParentVersion(parent string) ([]string, error) {
	r := regexp.MustCompile("^projects/" + NameRegex +
		"/products/" + NameRegex +
		"/versions/" + NameRegex +
		"$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid version '%s'", parent)
	}
	return m[0], nil
}
