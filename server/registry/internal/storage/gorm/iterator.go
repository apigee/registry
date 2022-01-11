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
	"fmt"

	"github.com/apigee/registry/server/registry/internal/storage/models"
	"google.golang.org/api/iterator"
)

// Iterator can be used to iterate through results of a query.
type Iterator struct {
	values interface{}
	index  int
}

// Next gets the next value from the iterator.
func (it *Iterator) Next(v interface{}) error {
	switch x := v.(type) {
	case *models.Project:
		values := it.values.([]models.Project)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	case *models.Api:
		values := it.values.([]models.Api)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	case *models.Version:
		values := it.values.([]models.Version)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	case *models.Spec:
		values := it.values.([]models.Spec)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	case *models.Deployment:
		values := it.values.([]models.Deployment)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	case *models.Blob:
		values := it.values.([]models.Blob)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	case *models.Artifact:
		values := it.values.([]models.Artifact)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	case *models.SpecRevisionTag:
		values := it.values.([]models.SpecRevisionTag)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	case *models.DeploymentRevisionTag:
		values := it.values.([]models.DeploymentRevisionTag)
		if it.index < len(values) {
			*x = values[it.index]
			it.index++
			return nil
		}
		return iterator.Done
	default:
		return fmt.Errorf("unsupported iterator type: %t", v)
	}
}
