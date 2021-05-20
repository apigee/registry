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

	"github.com/apigee/registry/server/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type weight string

const (
	weightA weight = "A"
	weightB weight = "B"
	weightC weight = "C"
	weightD weight = "D"
)

type Lexeme struct {
	Key       string `gorm:"primaryKey"`
	Kind      string
	ProjectID string
	Vector    TSVector
	Raw       string // stores raw text for excerpting; has excerpt from search result
}

func (x Lexeme) IsEmpty() bool {
	return x.Vector.rawText == ""
}

func NewLexemeForAPI(api *Api) *Lexeme {
	text := api.Description
	return &Lexeme{
		Key:       api.Key,
		Kind:      storage.ApiEntityName,
		ProjectID: api.ProjectID,
		Vector:    TSVector{rawText: text, weight: weightB},
		Raw:       text,
	}
}

type TSVector struct {
	rawText string
	weight  weight
}

func (t TSVector) GormDataType() string {
	return "tsvector"
}

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
