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

	"github.com/apigee/registry/server/names"
)

const (
	// ProjectEntityName is the storage entity name for project resources.
	ProjectEntityName = "Project"
	// ApiEntityName is the storage entity name for API resources.
	ApiEntityName = "Api"
	// VersionEntityName is the storage entity name for API version resources.
	VersionEntityName = "Version"
	// SpecEntityName is the storage entity name for API spec resources.
	SpecEntityName = "Spec"
	// SpecRevisionTagEntityName is the storage entity name for API spec revision tag resources.
	SpecRevisionTagEntityName = "SpecRevisionTag"
	// ArtifactEntityName is the storage entity name for artifact resources.
	ArtifactEntityName = "Artifact"
)

type Client interface {
	Close()

	Get(ctx context.Context, k Key, v interface{}) error
	Put(ctx context.Context, k Key, v interface{}) (Key, error)
	Delete(ctx context.Context, k Key) error
	Run(ctx context.Context, q Query) Iterator

	IsNotFound(err error) bool
	NotFoundError() error

	NewKey(kind, name string) Key
	NewQuery(query string) Query

	DeleteChildrenOfProject(ctx context.Context, project names.Project) error
	DeleteChildrenOfApi(ctx context.Context, api names.Api) error
	DeleteChildrenOfVersion(ctx context.Context, version names.Version) error
	DeleteAllMatches(ctx context.Context, q Query) error
	DeleteChildrenOfSpec(ctx context.Context, spec names.Spec) error

	GetRecentSpecRevisions(ctx context.Context, offset int32, projectID, apiID, versionID string) Iterator
}

type Key interface {
	String() string
}

type Query interface {
	Require(name string, value interface{}) Query
	Descending(field string) Query
	ApplyOffset(int32) Query
}

type Iterator interface {
	Next(interface{}) (Key, error)
}
