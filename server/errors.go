package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// internalError ...
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
