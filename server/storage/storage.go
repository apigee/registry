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

package storage

// This is the data storage protocol currently used by the Registry server.

import (
	"context"

	"github.com/apigee/registry/server/models"
)

type Client interface {
	Close()

	Get(ctx context.Context, k Key, v interface{}) error
	Put(ctx context.Context, k Key, v interface{}) (Key, error)
	Delete(ctx context.Context, k Key) error
	Run(ctx context.Context, q Query) Iterator

	IsNotFound(err error) bool

	NewKey(kind, name string) Key
	NewQuery(query string) Query

	DeleteAllMatches(ctx context.Context, q Query) error
	DeleteChildrenOfProject(ctx context.Context, project *models.Project) error
	DeleteChildrenOfApi(ctx context.Context, api *models.Api) error
	DeleteChildrenOfVersion(ctx context.Context, version *models.Version) error
}

type Key interface {
	String() string
}

type Query interface {
	Filter(filter string, value interface{}) Query
	Require(name string, value interface{}) Query
	Order(order string) Query
	ApplyCursor(cursorStr string) (Query, error)
}

type Iterator interface {
	Next(interface{}) (Key, error)
	GetCursor(l int) (string, error)
}
