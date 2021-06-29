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
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/apigee/registry/rpc"
)

var defaultServer *RegistryServer
var dname string

func init() {
	var err error
	dname, err = ioutil.TempDir("", "testdb-")
	if err != nil {
		panic(err)
	}

	defaultServer = &RegistryServer{
		// to allow the server tests to run with Postgres
		// (and other backends including CloudSQL Postgres),
		// these can be set using the environment variables
		// that are defined in config/registry.yaml and
		// (if undefined), default to the following values:
		database:     "sqlite3",
		dbConfig:     fmt.Sprintf("%s/registry.db", dname),
		loggingLevel: loggingError,
	}
}

func deleteAllProjects(s *RegistryServer) {
	ctx := context.Background()
	for {
		req := &rpc.ListProjectsRequest{PageSize: 1}
		rsp, err := s.ListProjects(ctx, req)
		if err != nil {
			panic(err)
		}
		if len(rsp.Projects) > 0 {
			req := &rpc.DeleteProjectRequest{
				Name: rsp.Projects[0].Name,
			}
			_, err := s.DeleteProject(ctx, req)
			if err != nil {
				panic(err)
			}
		} else {
			break
		}
	}
}

func defaultTestServer(t *testing.T) *RegistryServer {
	t.Helper()
	// if we comment out both of the following lines, tests fail
	// because tests leave state behind when they finish
	//os.Remove(fmt.Sprintf("%s/registry.db", dname))
	deleteAllProjects(defaultServer)
	// ideally (?) tests would clean up after themselves
	// but the above could also be done in common teardown code
	return defaultServer
}
