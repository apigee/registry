// Copyright 2020 Google LLC. All Rights Reserved.

package names

import (
	"regexp"
)

// PropertiesRegexp returns a regular expression that matches collection of properties.
func PropertiesRegexp() *regexp.Regexp {
	return regexp.MustCompile(
		"^projects/" + NameRegex +
			"(/products/" + NameRegex +
			"(/versions/" + NameRegex +
			"(/specs/" + NameRegex +
			")?" +
			")?" +
			")?" +
			"/properties$")
}

// PropertyRegexp returns a regular expression that matches a property resource name.
func PropertyRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex +
		"(/products/" + NameRegex +
		"(/versions/" + NameRegex +
		"(/specs/" + NameRegex +
		")?" +
		")?" +
		")?" +
		"/properties/" + NameRegex + "$")
}
