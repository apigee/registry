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

// ApisRegexp returns a regular expression that matches collection of apis.
func ApisRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + identifier + "/apis$")
}

// ApiRegexp returns a regular expression that matches a api resource name.
func ApiRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + identifier + "/apis/" + identifier + "$")
}

// ParseApi parses the name of an Api.
func ParseApi(name string) ([]string, error) {
	r := ApiRegexp()
	m := r.FindStringSubmatch(name)
	if m == nil {
		return nil, fmt.Errorf("invalid api name %q: must match %q", name, r)
	}
	return m, nil
}
