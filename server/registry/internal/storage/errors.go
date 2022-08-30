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
	"context"
	"net"
	"os"

	"github.com/apigee/registry/log"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func alreadyExists(err error) bool {
	// handle sqlite3 errors separately so their support can be conditionally compiled.
	if isSQLite3ErrorAlreadyExists(err) {
		return true
	}
	switch v := err.(type) {
	case *pq.Error:
		if v.Code.Name() == "unique_violation" {
			return true
		}
	}
	return false
}

// grpcErrorForDBError converts recognized database error codes to grpc error codes.
func grpcErrorForDBError(ctx context.Context, err error) error {
	// if this error already has a gRPC status code, just return it
	if _, ok := status.FromError(err); ok {
		return err
	}
	if err == context.DeadlineExceeded {
		return status.Error(codes.DeadlineExceeded, err.Error())
	}
	// handle sqlite3 errors separately so their support can be conditionally compiled.
	if err2 := grpcErrorForSQLite3Error(ctx, err); err2 != nil {
		return err2
	}
	// handle all other known error types.
	switch v := err.(type) {
	case *pq.Error:
		if v.Code.Name() == "unique_violation" {
			return status.Error(codes.AlreadyExists, err.Error())
		} else if v.Code.Name() == "too_many_connections" {
			return status.Error(codes.Unavailable, err.Error())
		} else if v.Code.Name() == "cannot_connect_now" {
			return status.Error(codes.FailedPrecondition, err.Error())
		} else if v.Code.Name() == "query_canceled" {
			return status.Error(codes.Canceled, err.Error())
		}
		log.Infof(ctx, "Unhandled %T %+v code=%s name=%s", v, v, v.Code, v.Code.Name())
	case *net.OpError:
		if v.Op == "dial" {
			// The database is overloaded.
			return status.Error(codes.Unavailable, err.Error())
		}
		switch vv := v.Unwrap().(type) {
		case *os.SyscallError:
			if vv.Syscall == "connect" {
				// The database is not running.
				return status.Error(codes.FailedPrecondition, err.Error())
			} else if vv.Syscall == "dial" {
				// The database is overloaded.
				return status.Error(codes.Unavailable, err.Error())
			} else if vv.Syscall == "socket" {
				// The database is overloaded.
				return status.Error(codes.Unavailable, err.Error())
			} else if vv.Syscall == "read" {
				// The connection was reset by peer.
				return status.Error(codes.Unavailable, err.Error())
			}
			log.Infof(ctx, "Unhandled %T %+v %s", vv, vv, vv.Syscall)
		case *net.DNSError:
			// DNS is overloaded.
			return status.Error(codes.Unavailable, err.Error())
		default:
			log.Infof(ctx, "Unhandled %T %+v", vv, vv)
		}
	default:
		if err.Error() == "EOF" {
			return status.Error(codes.Unavailable, err.Error())
		}
		if err.Error() == "driver: bad connection" {
			return status.Error(codes.Unavailable, err.Error())
		}
		if err.Error() == "sql: statement is closed" {
			return status.Error(codes.Unavailable, err.Error())
		}
		if err.Error() == "context canceled" {
			return status.Error(codes.Canceled, err.Error())
		}
		log.Infof(ctx, "Unhandled %T %+v", err, err)
	}

	// All unrecognized codes fall through to become "Internal" errors.
	return status.Error(codes.Internal, err.Error())
}
