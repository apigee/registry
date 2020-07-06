package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const topicName = "changes"

func main() {

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "flame-eval")
	if err != nil {
		panic(err)
	}
	{
		topic, err := client.CreateTopic(context.Background(), topicName)
		if err != nil {
			code := status.Code(err)
			if code == codes.AlreadyExists {
				fmt.Printf("already exists\n")
			} else {
				panic(err)
			}
		} else {
			log.Printf("%+v", topic)
		}
	}
	topic := client.Topic(topicName)
	defer topic.Stop()
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
