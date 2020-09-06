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
	"log"
	"testing"
	"time"

	"github.com/apigee/registry/server/models"
)

func TestCRUD(t *testing.T) {
	ctx := context.TODO()

	c, _ := NewClient(ctx, "sqlite3", "/tmp/testing.db")
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
	//log.Printf("%+v", project2)
}

func TestLoad(t *testing.T) {

	ctx := context.TODO()

	c, _ := NewClient(ctx, "sqlite3", "/tmp/testing.db")
	c.reset()
	c.Close()

	var err error
	for i := 0; i < 9999; i++ {
		{
			done := make(chan bool, 1)
			go func(done chan bool) {
				c, err = NewClient(ctx, "sqlite3", "/tmp/testing.db")
				if err != nil {
					t.Fatalf("Unable to create client: %+v", err)
				}
				now := time.Now()
				apiID := fmt.Sprintf("api-%04d", i)
				log.Printf("%s", apiID)
				api := &models.Api{
					ProjectID:   "demo",
					ApiID:       apiID,
					Description: "Demonstration API",
					CreateTime:  now,
					UpdateTime:  now,
				}
				k := c.NewKey(models.ApiEntityName, api.ResourceName())
				// fail if api already exists
				existingApi := &models.Api{}
				err := c.Get(ctx, k, existingApi)
				if err == nil {
					t.Errorf(err.Error())
				}
				_, err = c.Put(ctx, k, api)
				if err != nil {
					t.Errorf(err.Error())
				}
				c.Close()

				done <- true
			}(done)
			<-done
		}
	}
}
