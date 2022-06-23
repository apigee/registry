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

package registry

import (
	"context"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDuplicateCreation(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	db, err := server.getStorageClient(ctx)
	if err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	defer db.Close()

	// Create a project.
	project := models.NewProject(names.Project{ProjectID: "duplicate"}, &rpc.Project{})
	if err := db.CreateProject(ctx, project); err != nil {
		t.Errorf("error creating project %s", err)
	}
	// Create the project again and verify that it already exists.
	if err := db.CreateProject(ctx, project); status.Code(err) != codes.AlreadyExists {
		t.Errorf("error creating duplicate project, code should be AlreadyExists but is instead %s", err)
	}

	// Note that the following creations don't create parents.
	// Saving to the database does not require that parents exist.

	// Create an API.
	api, err := models.NewApi(names.Api{ApiID: "duplicate"}, &rpc.Api{})
	if err != nil {
		t.Errorf("error creating api %s", err)
	}
	if err := db.CreateApi(ctx, api); err != nil {
		t.Errorf("error creating api %s", err)
	}
	// Create the API again and verify that the response shows that it it already exists.
	if err := db.CreateApi(ctx, api); status.Code(err) != codes.AlreadyExists {
		t.Errorf("error creating duplicate api, code should be AlreadyExists but is instead %s", err)
	}

	// Create a version.
	version, err := models.NewVersion(names.Version{VersionID: "duplicate"}, &rpc.ApiVersion{})
	if err != nil {
		t.Errorf("error creating version %s", err)
	}
	if err := db.CreateVersion(ctx, version); err != nil {
		t.Errorf("error creating version %s", err)
	}
	// Create the version again and verify that the response shows that it already exists.
	if err := db.CreateVersion(ctx, version); status.Code(err) != codes.AlreadyExists {
		t.Errorf("error creating duplicate version, code should be AlreadyExists but is instead %s", err)
	}
}
