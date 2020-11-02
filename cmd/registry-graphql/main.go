// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"

	registry "github.com/apigee/registry/cmd/registry-graphql/graphql"
	"github.com/graphql-go/handler"
)

func main() {
	// graphql handler
	h := handler.New(&handler.Config{
		Schema: &registry.Schema,
		Pretty: true,
	})
	http.Handle("/graphql", h)

	// static file server for Graphiql in-browser editor
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// run the server
	port := "8088"
	fmt.Println("Running server on port " + port)
	http.ListenAndServe(":"+port, nil)
}
