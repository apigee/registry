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
	"context"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/golang/protobuf/ptypes/empty"
)

func TestServerIsRunning(t *testing.T) {
	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Get the server status.
	response, err := registryClient.GetStatus(ctx, &empty.Empty{})
	if err != nil {
		t.Logf("Failed to get status: %+v", err)
		t.FailNow()
	}
	if response.Message != "running" {
		t.Errorf("Invalid status response: %s", response.Message)
	}
}
