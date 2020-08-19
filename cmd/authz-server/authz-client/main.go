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
	"os"
	"time"

	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	"google.golang.org/grpc"
)

func main() {
	// Make a raw gRPC connection to a local authz-server.
	address := "localhost:50051"
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	// Create an Authorization API client from the connection.
	client := auth.NewAuthorizationClient(conn)

	// Get the auth token from the environment.
	token := os.Getenv("APG_REGISTRY_TOKEN")

	// Put the auth token in the headers that get sent with the CheckRequest.
	headers := make(map[string]string, 0)
	headers["authorization"] = "Bearer " + token

	// Build the CheckRequest.
	req := &auth.CheckRequest{
		Attributes: &auth.AttributeContext{
			Request: &auth.AttributeContext_Request{
				Http: &auth.AttributeContext_HttpRequest{
					Headers: headers,
				},
			},
		},
	}

	// Call the Check method.
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	res, err := client.Check(ctx, req)
	if res != nil {
		fmt.Printf("%+v\n", res)
	} else {
		log.Printf("Error %+v", err)
	}
}
