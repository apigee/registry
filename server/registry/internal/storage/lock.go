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

package storage

import (
	"context"
	"fmt"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
)

func (c *Client) lockTable(ctx context.Context, name string) *Client {
	// The LOCK TABLE statement below is unavailable in SQLite.
	// For SQLite, we instead lock all mutating transactions. See RegistryServer.runInTransaction().
	if c.DatabaseName(ctx) == "sqlite" {
		return c
	}
	return &Client{db: c.db.Exec(fmt.Sprintf("LOCK TABLE %s IN ACCESS EXCLUSIVE MODE", name))}
}

func (c *Client) LockProjects(ctx context.Context) *Client {
	return c.lockTable(ctx, "projects")
}

func (c *Client) LockApis(ctx context.Context) *Client {
	return c.lockTable(ctx, "apis")
}

func (c *Client) LockVersions(ctx context.Context) *Client {
	return c.lockTable(ctx, "versions")
}

func (c *Client) LockDeployments(ctx context.Context) *Client {
	return c.lockTable(ctx, "deployments")
}

func (c *Client) LockSpecs(ctx context.Context) *Client {
	return c.lockTable(ctx, "specs")
}

func (c *Client) LockArtifacts(ctx context.Context) *Client {
	return c.lockTable(ctx, "artifacts")
}
