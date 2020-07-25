// Copyright 2020 Google LLC. All Rights Reserved.

package names

import (
	"fmt"
	"regexp"
)

// VersionsRegexp returns a regular expression that matches a collection of versions.
func VersionsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex + "/products/" + NameRegex + "/versions$")
}

// VersionRegexp returns a regular expression that matches a version resource name.
func VersionRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex + "/products/" + NameRegex + "/versions/" + NameRegex + "$")
}

// ParseParentProduct ...
func ParseParentProduct(parent string) ([]string, error) {
	r := regexp.MustCompile("^projects/" + NameRegex +
		"/products/" + NameRegex +
		"$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	return m[0], nil
}
