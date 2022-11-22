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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// This makes a raw gRPC connection.
	// see the connection package for a simpler way to get a Go client.
	var opts []grpc.DialOption
	if os.Getenv("APG_REGISTRY_INSECURE") != "" {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			log.Fatal("failed to load system root CA cert pool")
		}
		creds := credentials.NewTLS(&tls.Config{
			RootCAs: systemRoots,
		})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	address := os.Getenv("APG_REGISTRY_ADDRESS")
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	// Create a Registry API client from the connection.
	client := rpc.NewRegistryClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configure the context to use an auth token from the environment.
	token := os.Getenv("APG_REGISTRY_TOKEN")
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	// Make a sample gRPC API call.
	req := &rpc.ListApisRequest{
		Parent: "projects/-",
	}
	res, err := client.ListApis(ctx, req)
	if res != nil {
		fmt.Println("The names of your apis:")
		for _, api := range res.Apis {
			fmt.Println(api.Name)
		}
	} else {
		log.Printf("Error %+v", err)
	}
}
