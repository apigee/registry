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

package server

import "cloud.google.com/go/datastore"

func boundPageSize(pageSize int32) int {
	if pageSize > 1000 {
		return 1000
	}
	if pageSize <= 0 {
		return 50
	}
	return int(pageSize)
}

func queryApplyPageSize(q *datastore.Query, pageSize int32) *datastore.Query {
	return q.Limit(boundPageSize(pageSize))
}

func queryApplyCursor(q *datastore.Query, cursorStr string) (*datastore.Query, error) {
	if cursorStr != "" {
		cursor, err := datastore.DecodeCursor(cursorStr)
		if err != nil {
			return nil, internalError(err)
		}
		q = q.Start(cursor)
	}
	return q, nil
}

// Get the cursor for the next page of results.
func iteratorGetCursor(it *datastore.Iterator, l int) (string, error) {
	if l > 0 {
		nextCursor, err := it.Cursor()
		if err != nil {
			return "", internalError(err)
		}
		return nextCursor.String(), nil
	}
	return "", nil
}
