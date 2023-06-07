// Copyright 2021 Google LLC.
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

package interceptor

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/apigee/registry/pkg/log"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	resourceOperation   interface{ GetName() string }
	collectionOperation interface{ GetParent() string }
)

// CallLogger returns a gRPC server interceptor for logging API operations.
func CallLogger(opts ...log.Option) grpc.UnaryServerInterceptor {
	// Create a logger scoped to this interceptor, configured with the provided options.
	// Each request will share this logger as a base template.
	sharedLogger := log.NewLogger(opts...)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		reqInfo := map[string]interface{}{
			"request_id": fmt.Sprintf("%.8s", uuid.New()),
			"method":     filepath.Base(info.FullMethod),
		}

		if r, ok := req.(resourceOperation); ok {
			reqInfo["resource"] = r.GetName()
		}

		if r, ok := req.(collectionOperation); ok {
			reqInfo["parent"] = r.GetParent()
		}

		// Bind request-scoped and inbound attributes to the context logger before handling the request.
		logger := log.WithInboundFields(ctx, sharedLogger).WithFields(reqInfo)
		ctx = log.NewContext(ctx, logger)

		logger.Info("Handling request.")
		start := time.Now()
		resp, err := handler(ctx, req)

		respInfo := map[string]interface{}{
			"duration":    time.Since(start),
			"status_code": status.Code(err),
		}

		if r, ok := resp.(resourceOperation); err == nil && ok {
			respInfo["resource"] = r.GetName()
		}

		// Bind response details before logging a response.
		logger = logger.WithFields(respInfo)

		// Error messages may include a status code, but we want to log messages and codes separately.
		if err != nil {
			st, _ := status.FromError(err)
			unwrapped := errors.New(st.Message())
			logger = logger.WithError(unwrapped)
		}

		switch status.Code(err) {
		case codes.OK:
			logger.Info("Success.")
		case codes.Canceled:
			logger.Info("Canceled.")
		case codes.Unknown:
			logger.Error("Unknown error.")
		case codes.InvalidArgument:
			logger.Error("Invalid argument.")
		case codes.DeadlineExceeded:
			logger.Error("Deadline exceeded.")
		case codes.NotFound:
			logger.Info("Not found.")
		case codes.AlreadyExists:
			logger.Error("Already exists.")
		case codes.PermissionDenied:
			logger.Error("Permission denied.")
		case codes.ResourceExhausted:
			logger.Error("Resource exhausted.")
		case codes.FailedPrecondition:
			logger.Error("Failed precondition.")
		case codes.Aborted:
			logger.Error("Aborted.")
		case codes.OutOfRange:
			logger.Error("Out of range.")
		case codes.Unimplemented:
			logger.Error("Unimplemented.")
		case codes.Internal:
			logger.Error("Internal error.")
		case codes.Unavailable:
			logger.Info("Unavailable.")
		case codes.DataLoss:
			logger.Info("Data loss.")
		case codes.Unauthenticated:
			logger.Info("Unauthenticated.")
		default:
			logger.Info("User error.")
		}

		return resp, err
	}
}
