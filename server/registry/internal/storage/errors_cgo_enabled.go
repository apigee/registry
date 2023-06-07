// Copyright 2022 Google LLC.
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

//go:build cgo

package storage

import (
	"context"

	"github.com/apigee/registry/pkg/log"
	"github.com/mattn/go-sqlite3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// returns true if the error is from sqlite3 and represents an "already exists" error.
func isSQLite3ErrorAlreadyExists(err error) bool {
	switch v := err.(type) {
	case sqlite3.Error:
		return v.Code == sqlite3.ErrNo(sqlite3.ErrConstraint) && v.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
	default:
		return false
	}
}

// returns an appropriate gRPC error code for a sqlite3 error.
func grpcErrorForSQLite3Error(ctx context.Context, err error) error {
	switch v := err.(type) {
	case sqlite3.Error:
		if v.Code == sqlite3.ErrNo(sqlite3.ErrConstraint) && v.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			return status.Error(codes.AlreadyExists, err.Error())
		}
		if v.Code == sqlite3.ErrNo(sqlite3.ErrBusy) ||
			v.Code == sqlite3.ErrNo(sqlite3.ErrCantOpen) ||
			v.Code == sqlite3.ErrNo(sqlite3.ErrReadonly) {
			return status.Error(codes.Unavailable, err.Error())
		}
		if v.Code == sqlite3.ErrConstraint && v.ExtendedCode == sqlite3.ErrConstraintForeignKey {
			return status.Error(codes.NotFound, v.ExtendedCode.Error())
		}
		log.Infof(ctx, "Unhandled %T %+v code=%d extended=%d", v, v, v.Code, v.ExtendedCode)
	}
	return nil
}
