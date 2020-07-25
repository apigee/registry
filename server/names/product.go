// Copyright 2020 Google LLC. All Rights Reserved.

package names

import (
	"fmt"
	"regexp"
)

// ProductsRegexp returns a regular expression that matches collection of products.
func ProductsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex + "/products$")
}

// ProductRegexp returns a regular expression that matches a product resource name.
func ProductRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex + "/products/" + NameRegex + "$")
}

// ParseParentProject ...
func ParseParentProject(parent string) ([]string, error) {
	r := regexp.MustCompile("^projects/" + NameRegex + "$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid project '%s'", parent)
	}
	return m[0], nil
}
