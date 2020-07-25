// Copyright 2020 Google LLC. All Rights Reserved.

package names

import (
	"regexp"
)

// LabelsRegexp returns a regular expression that matches collection of labels.
func LabelsRegexp() *regexp.Regexp {
	return regexp.MustCompile(
		"^projects/" + NameRegex +
			"(/products/" + NameRegex +
			"(/versions/" + NameRegex +
			"(/specs/" + NameRegex +
			")?" +
			")?" +
			")?" +
			"/labels$")
}

// LabelRegexp returns a regular expression that matches a Label resource name.
func LabelRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + NameRegex +
		"(/products/" + NameRegex +
		"(/versions/" + NameRegex +
		"(/specs/" + NameRegex +
		")?" +
		")?" +
		")?" +
		"/labels/" + NameRegex + "$")
}
