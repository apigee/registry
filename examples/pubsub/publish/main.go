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
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/apigee/registry/server/registry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	// Publish test messages on the current project.
	projectID := os.Getenv("REGISTRY_PROJECT_IDENTIFIER")

	// Create a pubsub client.
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	// Create the topic (or use it if it already exists).
	topic, err := client.CreateTopic(ctx, registry.TopicName)
	if err != nil {
		code := status.Code(err)
		if code == codes.AlreadyExists {
			topic = client.Topic(registry.TopicName)
		} else {
			panic(err)
		}
	}
	defer topic.Stop()

	// Publish a sample message.
	// Ordinarily, we would expect only the server to post on this topic.
	var results []*pubsub.PublishResult
	r := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("hello world"),
	})
	results = append(results, r)
	for _, r := range results {
		id, err := r.Get(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Published a message with a message ID: %s\n", id)
	}
}
