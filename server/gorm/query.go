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

	"github.com/apigee/registry/server/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Query represents a query in a storage provider.
type Query struct {
	Kind         string
	Cursor       string
	Limit        int
	Requirements []*Requirement
}

// Requirement adds an equality filter to a query.
type Requirement struct {
	Name  string
	Value interface{}
}

// NewQuery creates a new query.
func (c *Client) NewQuery(kind string) storage.Query {
	return &Query{
		Kind:   kind,
		Cursor: "",
		Limit:  50,
	}
}

// internalError ...
func internalError(err error) error {
	if err == nil {
		return nil
	}
	// TODO: selectively mask error details depending on caller privileges
	return status.Error(codes.Internal, err.Error())
}

// Filter adds a general filter to a query.
func (q *Query) Filter(name string, value interface{}) storage.Query {
	log.Fatalf("Unsupported filter %s %+v", name, value)
	return q
}

// Require adds a filter to a query that requires a field to have a specified value.
func (q *Query) Require(name string, value interface{}) storage.Query {
	switch name {
	case "ProjectID":
		name = "project_id"
	case "ApiID":
		name = "api_id"
	case "VersionID":
		name = "version_id"
	case "SpecID":
		name = "spec_id"
	case "Currency":
		name = "currency"
	default:
		log.Fatalf("UNEXPECTED REQUIRE TYPE: %s", name)
	}
	q.Requirements = append(q.Requirements, &Requirement{Name: name, Value: value})
	return q
}

func (q *Query) Order(value string) storage.Query {
	// ordering is ignored
	return q
}

// ApplyCursor configures a query to start from a specified cursor.
func (q *Query) ApplyCursor(cursorStr string) (storage.Query, error) {
	q.Cursor = cursorStr
	return q, nil
}
