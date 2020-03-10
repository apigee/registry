// Copyright 2019 Google Inc. All Rights Reserved.
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

	rpc "apigov.dev/flame/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func main() {

	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal("failed to load system root CA cert pool")
	}
	creds := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))

	address := os.Getenv("CLI_FLAME_ADDRESS")
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := rpc.NewFlameClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	token := os.Getenv("CLI_FLAME_TOKEN")
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	req := &rpc.ListProductsRequest{}
	req.Parent = "projects/google"
	res, err := client.ListProducts(ctx, req)
	if res != nil {
		fmt.Println("The names of your products:")
		for _, product := range res.Products {
			fmt.Println(product.Name)
		}
	} else {
		log.Printf("Error %+v", err)
	}
}
