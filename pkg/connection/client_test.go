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

package connection

import (
	"context"
	"testing"

	"github.com/apigee/registry/pkg/config/test"
)

func TestClientBadConfig(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))
	t.Setenv("REGISTRY_ADDRESS", "")
	t.Setenv("REGISTRY_INSECURE", "")

	_, err := NewRegistryClient(context.Background())
	if err == nil {
		t.Errorf("expected error")
	}

	_, err = NewRegistryClientWithSettings(context.Background(), Config{})
	if err == nil {
		t.Errorf("expected error")
	}

	_, err = NewAdminClient(context.Background())
	if err == nil {
		t.Errorf("expected error")
	}
	_, err = NewAdminClientWithSettings(context.Background(), Config{})
	if err == nil {
		t.Errorf("expected error")
	}

	_, err = NewProvisioningClient(context.Background())
	if err == nil {
		t.Errorf("expected error")
	}
	_, err = NewProvisioningClientWithSettings(context.Background(), Config{})
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestClientGoodConfig(t *testing.T) {
	t.Cleanup(test.CleanConfigDir(t))
	t.Setenv("REGISTRY_ADDRESS", "localhost:8080")
	t.Setenv("REGISTRY_INSECURE", "true")

	_, err := NewRegistryClient(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = NewAdminClient(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = NewProvisioningClient(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
