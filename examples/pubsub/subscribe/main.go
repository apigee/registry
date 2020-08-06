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
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
)

func main() {
	// Listen for changes on the current project.
	projectID := os.Getenv("REGISTRY_PROJECT_IDENTIFIER")
	topicName := "changes"
	subscriptionName := "my-subscription"

	// Create a pubsub client and subscription.
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}
	topic := client.Topic(topicName)
	defer topic.Stop()
	sub, err := client.CreateSubscription(
		ctx,
		subscriptionName,
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err == nil {
		log.Printf("created new subscription %+v", sub)
	} else {
		// assume this error occurred because the subscription already exists.
		sub = client.Subscription(subscriptionName)
	}

	// Create a cancelable context so that we can stop listening.
	cctx, cancel := context.WithCancel(ctx)

	// Start the listener.
	go func() {
		err = sub.Receive(
			cctx,
			func(ctx context.Context, m *pubsub.Message) {
				log.Printf("%s", string(m.Data))
				m.Ack()
			})
		if err != nil {
			panic(err)
		}
	}()

	// Listen for a while.
	time.Sleep(600 * time.Second)

	// Cancel the subscription and exit.
	cancel()
	err = sub.Delete(context.Background())
	if err != nil {
		panic(err)
	}
}
