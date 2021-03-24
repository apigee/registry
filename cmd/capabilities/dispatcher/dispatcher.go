// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
        "log"
        "time"
        "os"
        "context"
        "runtime"
        "regexp"
        "encoding/json"
        "cloud.google.com/go/pubsub"
        "github.com/apigee/registry/rpc"
	    "github.com/golang/protobuf/jsonpb"
        "google.golang.org/grpc/codes"
        "google.golang.org/grpc/status"
        "github.com/apigee/registry/cmd/capabilities/utils"
        cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
        taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

const topicName = "changes"
const subscriptionName = topicName + "-pull-subscriber"

func main() {
    log.Print("Starting subscriber...")
    ctx := context.Background()

    // Setup and start the dispatcher server
    dispatcher := &Dispatcher{}
    if err := dispatcher.setUp(ctx); err != nil {
        log.Printf(err.Error())
        return
    }

    if err := dispatcher.startServer(ctx); err != nil {
        log.Printf(err.Error())
    }
    return
}

type Dispatcher struct {
    pubsubClient *pubsub.Client
    subscription *pubsub.Subscription
}

func (d *Dispatcher) setUp(ctx context.Context) error {
    var err error
    d.pubsubClient, err = pubsub.NewClient(ctx, os.Getenv("REGISTRY_PROJECT_IDENTIFIER"))
    if err != nil {
        return err
    }

    var topic *pubsub.Topic
    topic, err = d.pubsubClient.CreateTopic(ctx, topicName)
    if status.Code(err) == codes.AlreadyExists {
        topic = d.pubsubClient.Topic(topicName)
    } else if err != nil {
        return err
    }

    d.subscription, err = d.pubsubClient.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
            Topic:            topic,
            AckDeadline:      10 * time.Second,
        })
    if status.Code(err) == codes.AlreadyExists {
        d.subscription = d.pubsubClient.Subscription(subscriptionName)
    } else if err != nil {
        return err
    }

    log.Printf("Created subscription: %s", d.subscription)

    // Configure subscriber.
    d.subscription.ReceiveSettings.MaxOutstandingMessages = 10
    d.subscription.ReceiveSettings.NumGoroutines = runtime.NumCPU()
    return nil
}

func (d *Dispatcher) startServer(ctx context.Context) error {
    err := d.subscription.Receive(ctx, messageHandler)
    if err != nil {
        return err
    }
    return nil
}

func messageHandler(ctx context.Context, msg *pubsub.Message) {
    data := string(msg.Data)
    message := rpc.Notification{}
    if err := jsonpb.UnmarshalString(data, &message); err != nil {
        log.Printf("Message data: %s", data)
        log.Printf("Error in json.Unmarshal: %v", err)
        msg.Ack()
        return
    }

    // Regex for specs
    r := `projects/.+/apis/.+/versions/.+/specs/.+`
    matched, err := regexp.Match(r, []byte(message.Resource))
    if err!= nil {
        log.Printf("Error parsing regex: %s", err)
        // Nack message so that it will be retried by pub/sub
        msg.Nack()
        return
    }

    // Create a task only if the resource is a spec
    if !matched {
        log.Printf("Resource is not a spec: %s", message.Resource)
        msg.Ack()
        return
    }

    switch changeType := message.Change; changeType {
        case rpc.Notification_CREATED, rpc.Notification_UPDATED:
            log.Printf("Creating task for change type %q", changeType)
        default:
            log.Printf("Ignoring change type %q", changeType)
            msg.Ack()
            return
    }

    if err := createQueueTask(ctx, message.Resource); err!= nil {
        log.Printf("Error creating queue task: %s", err)
        msg.Nack()
        return
    }

    msg.Ack()
    return
}

func createQueueTask(ctx context.Context, resource string) error {
    // Create CloudTasks client
    client, err := cloudtasks.NewClient(ctx)
    if err != nil {
            return err
    }
    // Build the Task queue path.
    queuePath := os.Getenv("TASK_QUEUE_ID")
    workerUrl := os.Getenv("WORKER_URL")

    // Build the request body
    body := utils.WorkerRequest{
        Resource: resource,
        Command: "registry compute lint",
    }
    jsonBody, err := json.Marshal(body)
    if err != nil {
        return err
    }

    req := &taskspb.CreateTaskRequest{
        Parent: queuePath,
        Task: &taskspb.Task{
            // https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#HttpRequest
            MessageType: &taskspb.Task_HttpRequest{
                HttpRequest: &taskspb.HttpRequest{
                    HttpMethod: taskspb.HttpMethod_POST,
                    Url:        workerUrl,
                    Headers:    map[string]string{"content-type": "application/json"},
                    Body:       []byte(jsonBody),
                },
            },
        },
    }

    createdTask, err := client.CreateTask(ctx, req)
    if err != nil {
            return err
    }

    log.Printf("Created task: %v", createdTask)
    return nil
}