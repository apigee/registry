package main

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
)

const topicName = "changes"

func main() {

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "flame-eval")
	if err != nil {
		panic(err)
	}

	topic := client.Topic(topicName)
	defer topic.Stop()

	sub, err := client.CreateSubscription(
		context.Background(),
		"sub",
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err != nil {
		sub = client.Subscription("sub")
	} else {
		log.Printf("created new subscription %+v", sub)
	}

	cctx, cancel := context.WithCancel(ctx)

	go func() {
		err = sub.Receive(
			cctx,
			func(ctx context.Context, m *pubsub.Message) {
				log.Printf("Got message: %+v", m)
				log.Printf("%s", string(m.Data))
				m.Ack()
			})
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(600 * time.Second)

	cancel()
	err = sub.Delete(context.Background())
	if err != nil {
		panic(err)
	}
}
