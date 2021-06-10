// Copyright 2021 Google LLC. All Rights Reserved.
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

package core

import (
	"fmt"
)

// UpdateMap updates a map containing labels or annotations to be modified.
func UpdateMap(m map[string]string,
	keyOverwrite bool,
	keysToSet map[string]string,
	keysToClear []string) (map[string]string, error) {
	if m == nil {
		m = make(map[string]string)
	}
	if !keyOverwrite {
		for k, _ := range keysToSet {
			if v, ok := m[k]; ok {
				return nil, fmt.Errorf("%q already has a value (%s), and --overwrite is false", k, v)
			}
		}
	}
	for _, k := range keysToClear {
		delete(m, k)
	}
	for k, v := range keysToSet {
		m[k] = v
	}
	return m, nil
}