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

//go:build !cgo

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
	switch v := err.(type) {
	case *pq.Error:
		if v.Code.Name() == "unique_violation" {
			return true
		}
	}
	return false
}

// grpcErrorForDBError converts recognized database error codes to grpc error codes.
func grpcErrorForDBError(err error) error {
	if _, ok := status.FromError(err); ok {
		return err
	}
	switch v := err.(type) {
	case *pq.Error:
		if v.Code.Name() == "unique_violation" {
			return status.Error(codes.AlreadyExists, err.Error())
		} else if v.Code.Name() == "too_many_connections" {
			return status.Error(codes.Unavailable, err.Error())
		}
		log.Infof(context.TODO(), "Unhandled %T %+v code=%s name=%s", v, v, v.Code, v.Code.Name())
	case *net.OpError:
		switch vv := v.Unwrap().(type) {
		case *os.SyscallError:
			if vv.Syscall == "dial" {
				return status.Error(codes.Unavailable, err.Error())
			}
			if vv.Syscall == "socket" {
				return status.Error(codes.Unavailable, err.Error())
			}
			log.Infof(context.TODO(), "Unhandled %T %+v %s", vv, vv, vv.Syscall)
		case *net.DNSError:
			return status.Error(codes.Unavailable, err.Error())
		default:
			log.Infof(context.TODO(), "Unhandled %T %+v", vv, vv)
		}
	default:
		if err.Error() == "sql: statement is closed" {
			return status.Error(codes.Unavailable, err.Error())
		}
		log.Infof(context.TODO(), "Unhandled %T %+v", err, err)
	}

	// All unrecognized codes fall through to become "Internal" errors.
	return status.Error(codes.Internal, err.Error())
}
