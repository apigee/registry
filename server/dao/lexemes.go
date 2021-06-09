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

package dao

import (
	"context"
	"database/sql"

	"github.com/apigee/registry/server/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LexemeRows struct {
	Rows []models.Lexeme
}

func (s *LexemeRows) Append(rows *sql.Rows) error {
	for rows.Next() {
		var row models.Lexeme
		if err := rows.Scan(&row.Key, &row.Raw); err != nil {
			return err
		}
		s.Rows = append(s.Rows, row)
	}
	return nil
}

func (d *DAO) SaveLexemes(ctx context.Context, lexemes []*models.Lexeme) error {
	for _, x := range lexemes {
		if x.IsEmpty() {
			//TODO delete Lexeme
		} else {
			if err := d.SaveLexeme(ctx, x); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DAO) SaveLexeme(ctx context.Context, lexeme *models.Lexeme) error {
	k := d.NewKey(string(lexeme.Field), lexeme.Key)
	if _, err := d.Put(ctx, k, lexeme); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

const textSearchQuery = `
SELECT
  key, ts_headline(raw, q) as excerpt
FROM (
  SELECT
    key, raw, ts_rank(vector, q) as rank, q
  FROM
    lexemes, plainto_tsquery(?) q
  WHERE
    vector @@ q
  ORDER BY
    rank DESC
) AS subquery_because_ts_headline_is_expensive_and_subquery_needs_a_name
`

func (d *DAO) ListLexemes(ctx context.Context, query string) (*LexemeRows, error) {
	var rows LexemeRows
  if err := d.Raw(ctx, &rows, textSearchQuery, query); err != nil {
  	return nil, err
	}
	return &rows, nil
}
