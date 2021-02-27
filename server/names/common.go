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

	"github.com/google/uuid"
)

// The format of a resource identifier.
// This may be extended to include all characters that do not require escaping.
// See https://aip.dev/122#resource-id-segments.
const identifier = "([a-zA-Z0-9-_\\.]+)"

// The format of a custom revision tag.
const revisionTag = "(@[a-zA-z0-9-]+)?"

// GenerateID generates a random resource ID.
func GenerateID() string {
	return uuid.New().String()[:8]
}

// ValidateID returns an error if the provided ID is invalid.
func ValidateID(id string) error {
	r := regexp.MustCompile("^" + identifier + "$")
	m := r.FindStringSubmatch(id)
	if m == nil {
		return fmt.Errorf("invalid id %q, must match %q", id, r)
	}
	return nil
}
