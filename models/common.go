package models

import (
	"fmt"
	"regexp"
)

const nameRegex = "([a-zA-Z0-9-_\\.]+)"

func validateID(id string) error {
	r := regexp.MustCompile("^" + nameRegex + "$")
	m := r.FindAllStringSubmatch(id, -1)
	if m == nil {
		return fmt.Errorf("invalid id '%s'", id)
	}
	return nil
}

// ProductsRegexp returns a regular expression that matches collection of products.
func ProductsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products$")
}

// ProductRegexp returns a regular expression that matches a product resource name.
func ProductRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products/" + nameRegex + "$")
}

// VersionsRegexp returns a regular expression that matches a collection of versions.
func VersionsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products/" + nameRegex + "/versions$")
}

// VersionRegexp returns a regular expression that matches a version resource name.
func VersionRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products/" + nameRegex + "/versions/" + nameRegex + "$")
}

// SpecsRegexp returns a regular expression that matches a collection of specs.
func SpecsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products/" + nameRegex + "/versions/" + nameRegex + "/specs$")
}

// SpecRegexp returns a regular expression that matches a spec resource name.
func SpecRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products/" + nameRegex + "/versions/" + nameRegex + "/specs/" + nameRegex + "$")
}

// FilesRegexp returns a regular expression that matches a collection of files.
func FilesRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products/" + nameRegex + "/versions/" + nameRegex + "/specs/" + nameRegex + "/files$")
}

// FileRegexp returns a regular expression that matches a file resource name.
func FileRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/products/" + nameRegex + "/versions/" + nameRegex + "/specs/" + nameRegex + "/files/" + nameRegex + "$")
}
