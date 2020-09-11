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

// We might extend this to all characters that do not require escaping.
// See "Resource ID Segments" in https://aip.dev/122.
const NameRegex = "([a-zA-Z0-9-_\\.]+)"

// Generated revision names are lowercase hex strings, but we also
// allow user-specified revision tags which can be mixed-case strings
// containing dashes.
const RevisionRegex = "(@[a-zA-z0-9-]+)?"

func ValidateID(id string) error {
	r := regexp.MustCompile("^" + NameRegex + "$")
	m := r.FindStringSubmatch(id)
	if m == nil {
		return fmt.Errorf("invalid id '%s'", id)
	}
	return nil
}

func ValidateRevision(s string) error {
	r := regexp.MustCompile("^" + RevisionRegex + "$")
	m := r.FindStringSubmatch(s)
	if m == nil {
		return fmt.Errorf("invalid revision '%s'", s)
	}
	return nil
}
