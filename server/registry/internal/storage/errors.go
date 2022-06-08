// Copyright 2022 Google LLC. All Rights Reserved.
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

import (
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcErrorForDbError converts recognized database error codes to grpc error codes.
func grpcErrorForDbError(err error) error {
	switch v := err.(type) {
	case *pq.Error: // Postgres codes are from https://www.postgresql.org/docs/current/errcodes-appendix.html
		if v.Code == "23505" { // unique_violation
			return status.Error(codes.AlreadyExists, err.Error())
		}
	}
	// All unrecognized codes fall through to become "Internal" errors.
	return status.Error(codes.Internal, err.Error())
}
