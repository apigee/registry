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
	"strings"

	"github.com/google/uuid"
)

// The format of a resource identifier.
// This may be extended to include all characters that do not require escaping.
// See https://aip.dev/122#resource-id-segments.
const identifier = "([a-zA-Z0-9-_\\.]+)"

// The format of a custom revision tag.
const revisionTag = "(@[a-zA-z0-9-]+)?"

// ValidateCustomID returns a descriptive validation error if the user provided ID is invalid.
func ValidateCustomID(id string) error {
	r := regexp.MustCompile("^[a-z0-9][a-z0-9-]{2,61}[a-z0-9]$")
	if len(id) < 4 || 63 < len(id) {
		return fmt.Errorf("invalid id %q: must be 4-63 characters in length", id)
	} else if strings.HasPrefix(id, "-") || strings.HasSuffix(id, "-") {
		return fmt.Errorf("invalid id %q: must not begin or end with hyphen", id)
	} else if _, err := uuid.Parse(id); err == nil {
		return fmt.Errorf("invalid id %q: must not match UUID syntax", id)
	} else if !r.MatchString(id) {
		return fmt.Errorf("invalid id %q: must match %q", id, r)
	}

	return nil
}
