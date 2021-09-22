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
	"flag"
	"fmt"
	"net/http"

	"github.com/apigee/registry/cmd/registry-graphql/graphql"
	"github.com/graphql-go/handler"
)

var corsAllowOriginFlag *string

type corsProxy struct {
	h http.Handler
}

func (p *corsProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if *corsAllowOriginFlag != "" {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", *corsAllowOriginFlag)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(""))
			return
		}
	}
	p.h.ServeHTTP(w, r)
}

func main() {
	// allow all CORS requests (not for production use)
	corsAllowOriginFlag = flag.String("cors-allow-origin", "", "allow all CORS requests from the specified origin")
	flag.Parse()

	// graphql handler
	h := handler.New(&handler.Config{
		Schema: &graphql.Schema,
		Pretty: true,
	})
	http.Handle("/graphql", &corsProxy{h: h})

	// static file server for Graphiql in-browser editor
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// run the server
	port := "8088"
	fmt.Println("Running server on port " + port)
	_ = http.ListenAndServe(":"+port, nil)
}
