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

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
)

func main() {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	// Make a sample gRPC API call.
	req := &rpc.ListApisRequest{
		Parent: "projects/-",
	}
	it := client.ListApis(ctx, req)
	for {
		api, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalf("%s", err.Error())
		}
		fmt.Println(api.Name)
	}
}
