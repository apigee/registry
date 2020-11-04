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

package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// internalError returns an error that wraps an internal server error.
func internalError(err error) error {
	if err == nil {
		return nil
	}
	// TODO: selectively mask error details depending on caller privileges
	return status.Error(codes.Internal, err.Error())
}

func invalidArgumentError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.InvalidArgument, err.Error())
}

func unavailableError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Unavailable, err.Error())
}

func notFoundError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.NotFound, err.Error())
}
