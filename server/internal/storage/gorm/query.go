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
	"log"
)

// Query represents a query in a storage provider.
type Query struct {
	Kind         string
	Offset       int
	Order        string
	Requirements []*Requirement
}

// Requirement adds an equality filter to a query.
type Requirement struct {
	Name  string
	Value interface{}
}

// NewQuery creates a new query.
func (c *Client) NewQuery(kind string) *Query {
	return &Query{
		Kind: kind,
	}
}

// Require adds a filter to a query that requires a field to have a specified value.
func (q *Query) Require(name string, value interface{}) *Query {
	switch name {
	case "ProjectID":
		name = "project_id"
	case "ApiID":
		name = "api_id"
	case "VersionID":
		name = "version_id"
	case "SpecID":
		name = "spec_id"
	case "RevisionID":
		name = "revision_id"
	case "ArtifactID":
		name = "artifact_id"
	default:
		log.Fatalf("UNEXPECTED REQUIRE TYPE: %s", name)
	}
	q.Requirements = append(q.Requirements, &Requirement{Name: name, Value: value})
	return q
}

func (q *Query) Descending(field string) *Query {
	switch field {
	case "RevisionCreateTime":
		q.Order = "revision_create_time desc"
	}

	return q
}

func (q *Query) ApplyOffset(offset int32) *Query {
	q.Offset = int(offset)
	return q
}
