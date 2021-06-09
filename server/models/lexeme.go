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

package models

import (
	"context"
	"fmt"
	"html"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type field string

const (
	fieldDisplayName field = "displayname"
	fieldDescription field = "description"
	fieldParameters  field = "parameters"
	fieldMethods     field = "methods"
	fieldSchemas     field = "schemas"
)

type weight string

const (
	weightA weight = "A"
	weightB weight = "B"
	weightC weight = "C"
	weightD weight = "D"
)

// Lexeme represents two slightly different states for full text search.
// To store text for searching, it must be normalized into a vector with
// the Postgres function ts_vector. The raw text is also stored so that
// it can be highlighted as search results, as the result of a search
// query.
type Lexeme struct {
	Key       string `gorm:"primaryKey"`
	Kind      string
	Field     field
	ProjectID string
	Vector    TSVector
	Raw       string // stores raw text for excerpting; has excerpt from search result
	escaped   bool
}

// escape should be called after filling struct to HTML-escape the raw text
// in the Vector, and also then copies it to Raw.
func (x *Lexeme) escape() *Lexeme {
	if x == nil || x.escaped {
		return x
	}

	x.Vector.rawText = html.EscapeString(x.Vector.rawText)
	x.Raw = x.Vector.rawText
	x.escaped = true
	return x
}

// IsEmpty determines whether an update should write or delete the Lexeme.
func (x *Lexeme) IsEmpty() bool {
	return x.Vector.rawText == ""
}

// TSVector opaquely represents the write-only ts_vector, containing
// the should-be-escaped text to index and the search weight.
type TSVector struct {
	rawText string
	weight  weight
}

// GormDataType of TSVector is the Postgres column type "tsvector".
func (t TSVector) GormDataType() string {
	return "tsvector"
}

// GormValue of TSVector returns the Postgres expression to convert
// search text to a weighted vector.
func (t TSVector) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	w := t.weight
	if w == "" {
		w = weightD
	}
	return clause.Expr{
		SQL:  "setweight(to_tsvector(?), ?)",
		Vars: []interface{}{t.rawText, w},
	}
}

// Scan implements the sql.Scanner interface
func (t *TSVector) Scan(v interface{}) error {
	t.rawText = fmt.Sprintf("%+v", v)
	return nil
}
