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

package storage

import (
	"context"
	"net"
	"os"

	"github.com/apigee/registry/pkg/log"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	if status.Code(err) == codes.AlreadyExists {
		return true
	}
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
	// unwrap err, if wrapped
	cause := errors.Cause(err)

	// if this error already has a gRPC status code, just return it
	if _, ok := status.FromError(cause); ok {
		return cause
	}

	log.Debugf(ctx, "DBError %+v", err) // log the stack trace

	if cause == context.DeadlineExceeded {
		return status.Error(codes.DeadlineExceeded, cause.Error())
	}
	// handle sqlite3 errors separately so their support can be conditionally compiled.
	if err2 := grpcErrorForSQLite3Error(ctx, cause); err2 != nil {
		return err2
	}
	// handle all other known error types.
	switch v := cause.(type) {
	case *pq.Error:
		if v.Code.Name() == "unique_violation" {
			return status.Error(codes.AlreadyExists, cause.Error())
		} else if v.Code.Name() == "too_many_connections" {
			return status.Error(codes.Unavailable, cause.Error())
		} else if v.Code.Name() == "cannot_connect_now" {
			return status.Error(codes.FailedPrecondition, cause.Error())
		} else if v.Code.Name() == "query_canceled" {
			return status.Error(codes.Canceled, cause.Error())
		} else if v.Code.Name() == "foreign_key_violation" {
			return status.Error(codes.NotFound, cause.Error())
		}
		log.Infof(ctx, "Unhandled %T %+v code=%s name=%s", v, v, v.Code, v.Code.Name())
	case *net.OpError:
		if v.Op == "dial" {
			// The database is overloaded.
			return status.Error(codes.Unavailable, cause.Error())
		}
		switch vv := v.Unwrap().(type) {
		case *os.SyscallError:
			if vv.Syscall == "connect" {
				// The database is not running.
				return status.Error(codes.FailedPrecondition, cause.Error())
			} else if vv.Syscall == "dial" {
				// The database is overloaded.
				return status.Error(codes.Unavailable, cause.Error())
			} else if vv.Syscall == "socket" {
				// The database is overloaded.
				return status.Error(codes.Unavailable, cause.Error())
			} else if vv.Syscall == "read" {
				// The connection was reset by peer.
				return status.Error(codes.Unavailable, cause.Error())
			}
			log.Infof(ctx, "Unhandled %T %+v %s", vv, vv, vv.Syscall)
		case *net.DNSError:
			// DNS is overloaded.
			return status.Error(codes.Unavailable, cause.Error())
		default:
			log.Infof(ctx, "Unhandled %T %+v", vv, vv)
		}
	default:
		if cause.Error() == "EOF" {
			return status.Error(codes.Unavailable, cause.Error())
		}
		if cause.Error() == "driver: bad connection" {
			return status.Error(codes.Unavailable, cause.Error())
		}
		if cause.Error() == "sql: statement is closed" {
			return status.Error(codes.Unavailable, cause.Error())
		}
		if cause.Error() == "context canceled" {
			return status.Error(codes.Canceled, cause.Error())
		}
		if cause.Error() == "pq: Could not complete operation in a failed transaction" {
			return status.Error(codes.Aborted, cause.Error())
		}
		log.Infof(ctx, "Unhandled %T %+v", cause, cause)
	}

	// All unrecognized codes fall through to become "Internal" errors.
	return status.Error(codes.Internal, cause.Error())
}
