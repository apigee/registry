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

package log

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// Metadata provides fields for multi-process log organization.
type Metadata struct {
	// UID is a unique identifier for logs. It can be used to group sets of related logs.
	UID string
}

// Identifiers for logging fields stored in gRPC metadata.
const (
	uidKey = "logging_uid"
)

// NewOutboundContext returns a new context with the provided metadata attached.
// This can be used to preserve the provided fields across gRPC calls.
func NewOutboundContext(ctx context.Context, md Metadata) context.Context {
	if outbound, ok := metadata.FromOutgoingContext(ctx); ok {
		outbound.Set(uidKey, md.UID)
		return metadata.NewOutgoingContext(ctx, outbound)
	}

	return metadata.AppendToOutgoingContext(ctx, uidKey, md.UID)
}

// loggerKey is an unexported type used to attach loggers as context values.
type loggerKey struct{}

// NewContext returns a new context that can be used to retrieve the provided logger.
// This can be used to share a logger instance as the context passes through a single process.
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext returns a logger from the provided context.
// Options will be applied to a default configuration if the context doesn't have an associated logger.
func FromContext(ctx context.Context, opts ...Option) Logger {
	logger, ok := ctx.Value(loggerKey{}).(Logger)
	if !ok {
		logger = NewLogger(opts...)
	}

	// Include metadata that was added locally to the context but hasn't been sent yet.
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		return withMetadataFields(logger, md)
	}

	return logger
}

// WithInboundFields returns a new logger including fields from the incoming context's metadata.
func WithInboundFields(ctx context.Context, logger Logger) Logger {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return withMetadataFields(logger, md)
	}

	return logger
}

func withMetadataFields(l Logger, md metadata.MD) Logger {
	fields := make(map[string]interface{}, 1)
	if vals, ok := md[uidKey]; ok {
		// When we add metadata to the context we ensure it only has one value.
		fields["uid"] = vals[0]
	}

	return l.WithFields(fields)
}
