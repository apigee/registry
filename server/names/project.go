// Copyright 2020 Google LLC. All Rights Reserved.

package names

import (
	"regexp"
)

// ProjectsRegexp returns a regular expression that matches collection of projects.
func ProjectsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects$")
}

// ProjectRegexp returns a regular expression that matches a project resource name.
func ProjectRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex + "$")
}
