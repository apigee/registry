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

package benchmark

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/wipeout"
)

// The level of parallelism to use for wipeout.
// If this is too high, the test could be rate-limited.
const jobs = 10

var (
	projectID string
)

func init() {
	flag.StringVar(&projectID, "project_id", "bench", "registry project ID")
}

func root() string {
	return fmt.Sprintf("projects/%s/locations/global", projectID)
}

func setup(b *testing.B) (context.Context, connection.Client) {
	b.Helper()
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Unable to connect to registry server. Is it running?")
	}
	wipeout.Wipeout(ctx, client, projectID, jobs)
	return ctx, client
}

func teardown(ctx context.Context, b *testing.B, client connection.Client) {
	b.Helper()
	wipeout.Wipeout(ctx, client, projectID, jobs)
}
