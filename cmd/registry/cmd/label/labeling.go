// Copyright 2021 Google LLC.
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

package label

import (
	"fmt"
)

// Labeling represents a user-specified change to a set of labels or annotations.
// Note that this same structure is used for both labels and annotations.
type Labeling struct {
	Overwrite bool
	Set       map[string]string
	Clear     []string
}

// Apply applies a labeling to a map. It returns the modified map
// because it creates the map if nil is passed in.
func (l *Labeling) Apply(m map[string]string) (map[string]string, error) {
	if m == nil {
		m = make(map[string]string)
	}
	if !l.Overwrite {
		for k := range l.Set {
			if v, ok := m[k]; ok {
				return nil, fmt.Errorf("%q already has a value (%s), and --overwrite is false", k, v)
			}
		}
	}
	for _, k := range l.Clear {
		delete(m, k)
	}
	for k, v := range l.Set {
		m[k] = v
	}
	return m, nil
}
