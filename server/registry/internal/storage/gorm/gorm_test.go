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

package gorm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestFieldClearing(t *testing.T) {
	ctx := context.Background()

	c, err := NewClient(ctx, "sqlite3", t.TempDir()+"/testing.db")
	if err != nil {
		t.Fatalf("NewClient returned error: %s", err)
	}
	defer c.Close()
	if err := c.EnsureTables(); err != nil {
		t.Fatalf("EnsureTables returned error: %s", err)
	}

	original := &models.Project{
		ProjectID:   "my-project",
		Description: "My Project",
	}

	k := c.NewKey(ProjectEntityName, original.Name())
	if _, err := c.Put(ctx, k, original); err != nil {
		t.Fatalf("Setup: Put(%q, %+v) returned error: %s", k, original, err)
	}

	update := &models.Project{
		ProjectID:   original.ProjectID,
		Description: "",
	}

	if _, err := c.Put(ctx, k, update); err != nil {
		t.Fatalf("Put(%q, %+v) returned error: %s", k, update, err)
	}

	got := new(models.Project)
	if err := c.Get(ctx, k, got); err != nil {
		t.Fatalf("Get(%q) returned error: %s", k, err)
	}

	if !cmp.Equal(got, update, protocmp.Transform()) {
		t.Errorf("Get(%q) returned unexpected diff (-want +got):\n%s", k, cmp.Diff(update, got, protocmp.Transform()))
	}
}

func TestCRUD(t *testing.T) {
	ctx := context.Background()

	c, err := NewClient(ctx, "sqlite3", t.TempDir()+"/testing.db")
	if err != nil {
		t.Fatalf("NewClient returned error: %s", err)
	}
	defer c.Close()
	if err := c.EnsureTables(); err != nil {
		t.Fatalf("EnsureTables returned error: %s", err)
	}

	now := time.Now()
	project := &models.Project{
		ProjectID:   "demo",
		DisplayName: "Demo",
		Description: "Demonstration Project",
		CreateTime:  now,
		UpdateTime:  now,
	}
	k := c.NewKey("Project", "projects/demo")

	// Create a project.
	if _, err := c.Put(ctx, k, project); err != nil {
		t.Errorf(err.Error())
	}

	// Verify that the project exists.
	err = c.Get(ctx, k, project)
	if err != nil {
		t.Errorf(err.Error())
	}
	if project.ProjectID != "demo" {
		t.Errorf("project creation failed")
	}

	// Update the project.
	project.ProjectID = "updated"
	project.DisplayName = "Updated"
	_, err = c.Put(ctx, k, project)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Verify the project update.
	_ = c.Get(ctx, k, project)
	if project.ProjectID != "updated" {
		t.Errorf("Project update failed")
	}

	// Delete the project.
	q := c.NewQuery("Project").Require("ProjectID", project.ProjectID)
	err = c.Delete(ctx, q)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Verify the deletion.
	var project2 models.Project
	err = c.Get(ctx, k, &project2)
	if !c.IsNotFound(err) {
		t.Errorf("Project deletion failed")
	}
}

func TestLoad(t *testing.T) {
	ctx := context.Background()

	db := t.TempDir() + "/testing.db"
	c, err := NewClient(ctx, "sqlite3", db)
	if err != nil {
		t.Fatalf("NewClient returned error: %s", err)
	}
	defer c.Close()
	if err := c.EnsureTables(); err != nil {
		t.Fatalf("EnsureTables returned error: %s", err)
	}

	for i := 0; i < 99; i++ {
		c, err := NewClient(ctx, "sqlite3", db)
		if err != nil {
			t.Fatalf("Unable to create client: %+v", err)
		}
		now := time.Now()
		api := &models.Api{
			ProjectID:   "demo",
			ApiID:       fmt.Sprintf("api-%04d", i),
			Description: "Demonstration API",
			CreateTime:  now,
			UpdateTime:  now,
		}
		k := c.NewKey(ApiEntityName, api.Name())
		if err := c.Get(ctx, k, &models.Api{}); err == nil {
			t.Errorf("API %q already exists, expected not found", api.Name())
		}
		if _, err := c.Put(ctx, k, api); err != nil {
			t.Errorf(err.Error())
		}
		c.Close()
	}
}
