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

import (
	"encoding/base64"
	"fmt"

	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/storage"
	"google.golang.org/api/iterator"
)

// Iterator can be used to iterate through results of a query.
type Iterator struct {
	Client *Client
	Values interface{}
	Index  int
	Cursor string
}

// GetCursor gets the cursor for the next page of results.
func (it *Iterator) GetCursor(l int) (string, error) {
	encodedCursor := base64.StdEncoding.EncodeToString([]byte(it.Cursor))
	return encodedCursor, nil
}

// Next gets the next value from the iterator.
func (it *Iterator) Next(v interface{}) (storage.Key, error) {
	switch x := v.(type) {
	case *models.Project:
		values := it.Values.([]models.Project)
		if it.Index < len(values) {
			*x = values[it.Index]
			it.Cursor = x.Key
			it.Index++
			return it.Client.NewKey("Project", x.Key), nil
		}
		return nil, iterator.Done
	case *models.Api:
		values := it.Values.([]models.Api)
		if it.Index < len(values) {
			*x = values[it.Index]
			it.Cursor = x.Key
			it.Index++
			return it.Client.NewKey("Api", x.Key), nil
		}
		return nil, iterator.Done
	case *models.Version:
		values := it.Values.([]models.Version)
		if it.Index < len(values) {
			*x = values[it.Index]
			it.Cursor = x.Key
			it.Index++
			return it.Client.NewKey("Version", x.Key), nil
		}
		return nil, iterator.Done
	case *models.Spec:
		values := it.Values.([]models.Spec)
		if it.Index < len(values) {
			*x = values[it.Index]
			it.Cursor = x.Key
			it.Index++
			return it.Client.NewKey("Spec", x.Key), nil
		}
		return nil, iterator.Done
	case *models.Blob:
		values := it.Values.([]models.Blob)
		if it.Index < len(values) {
			*x = values[it.Index]
			it.Cursor = x.Key
			it.Index++
			return it.Client.NewKey("Blob", x.Key), nil
		}
		return nil, iterator.Done
	case *models.Artifact:
		values := it.Values.([]models.Artifact)
		if it.Index < len(values) {
			*x = values[it.Index]
			it.Cursor = x.Key
			it.Index++
			return it.Client.NewKey("Artifact", x.Key), nil
		}
		return nil, iterator.Done
	case *models.SpecRevisionTag:
		values := it.Values.([]models.SpecRevisionTag)
		if it.Index < len(values) {
			*x = values[it.Index]
			it.Cursor = x.Key
			it.Index++
			return it.Client.NewKey("SpecRevisionTag", x.Key), nil
		}
		return nil, iterator.Done
	default:
		return nil, fmt.Errorf("unsupported iterator type: %t", v)
	}
}
