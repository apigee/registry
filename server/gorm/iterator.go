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

package gorm

// Iterator can be used to iterate through results of a query.
type Iterator struct {
}

// GetCursor gets the cursor for the next page of results.
func (it *Iterator) GetCursor(l int) (string, error) {
	return "", nil
}

// Next gets the next value from the iterator.
func (it *Iterator) Next(v interface{}) (*Key, error) {
	return nil, nil
}
