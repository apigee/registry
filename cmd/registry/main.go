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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/apigee/registry/cmd/registry/cmd"
	"github.com/apigee/registry/log"
	"github.com/google/uuid"
)

func main() {
	// Bind a logger instance to the local context with metadata for outbound requests.
	logger := log.NewLogger(log.DebugLevel)
	ctx := log.NewOutboundContext(log.NewContext(context.Background(), logger), log.Metadata{
		UID: fmt.Sprintf("%.8s", uuid.New()),
	})

	cmd := cmd.Command()
	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
