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

package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/apigee/registry/pkg/log"
	"google.golang.org/grpc/metadata"
)

func ExampleNewOutboundContext() {
	ctx := log.NewOutboundContext(context.Background(), log.Metadata{
		UID: "test_uid",
	})

	var (
		// Buffer to hold logs.
		buf bytes.Buffer
		// Logger that writes JSON entries to a buffer.
		logger = log.NewLogger(log.JSONFormat(&buf))
		// Struct to hold the parsed log entry.
		entry struct {
			Fields map[string]interface{} `json:"fields"`
		}
	)

	// "Call" a server that prints a log message.
	grpcFakeCall(ctx, func(ctx context.Context) {
		log.WithInboundFields(ctx, logger).Info("Print a server log with the unique ID")
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			panic(err)
		}

		// Output: Unique ID = test_uid
		fmt.Printf("Unique ID = %v", entry.Fields["uid"])
	})
}

// Converts outgoing gRPC metadata (if present) to incoming metadata before calling the handler.
func grpcFakeCall(ctx context.Context, handler func(context.Context)) {
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		ctx = metadata.NewIncomingContext(ctx, md) // Set incoming context with caller's outgoing metadata.
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{}) // Clear outgoing context.
	handler(ctx)
}
