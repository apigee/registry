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
	"log"
	"testing"
	"time"

	"github.com/apigee/registry/server/models"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func TestCRUD(t *testing.T) {
	ctx := context.TODO()

	c, _ := NewClient(ctx, "demo")
	defer c.Close()
	// delete and recreate database tables
	c.reset()

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
	_, err := c.Put(ctx, k, project)
	if err != nil {
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
	c.Get(ctx, k, project)
	if project.ProjectID != "updated" {
		t.Errorf("Project update failed")
	}

	// Delete the project.
	err = c.Delete(ctx, k)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Verify the deletion.
	var project2 models.Project
	err = c.Get(ctx, k, &project2)
	if !c.IsNotFound(err) {
		t.Errorf("Project deletion failed")
	}
	log.Printf("%+v", project2)

	c.Close()
}
