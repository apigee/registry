// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
