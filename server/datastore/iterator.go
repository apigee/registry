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

package datastore

import (
	"cloud.google.com/go/datastore"
	"github.com/apigee/registry/server/storage"
)

// Iterator can be used to iterate through results of a query.
type Iterator struct {
	iterator *datastore.Iterator
}

// Next gets the next item in the query results.
func (it *Iterator) Next(v interface{}) (storage.Key, error) {
	key, err := it.iterator.Next(v)
	if key == nil {
		return nil, err
	}
	return &Key{key: key}, err
}

// GetCursor gets the cursor for the next page of results.
func (it *Iterator) GetCursor() (string, error) {
	nextCursor, err := it.iterator.Cursor()
	if err != nil {
		return "", internalError(err)
	}
	return nextCursor.String(), nil
}
