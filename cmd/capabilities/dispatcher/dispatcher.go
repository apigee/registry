// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
        "log"
        "time"
//         "net/http"
        "os"
        "context"
        "runtime"
//         "fmt"
        "strings"
        "encoding/json"
        "cloud.google.com/go/pubsub"
        "github.com/apigee/registry/rpc"
	    "github.com/golang/protobuf/jsonpb"
        "google.golang.org/grpc/codes"
        "google.golang.org/grpc/status"

        cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
        taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

// Make this configurable
const topicName = "changes"
const subscriptionName = topicName + "-pull-subscriber"

func main() {
    log.Print("Starting subscriber...")
    ctx := context.Background()

    // Setup and start the dispatcher server
    dispatcher := &Dispatcher{}
    err := dispatcher.setUp(ctx)
    if err != nil {
        log.Printf(err.Error())
        return
    }
    err = dispatcher.startServer(ctx)
    if err != nil {
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
    // Create pubsub client
    d.pubsubClient, err = pubsub.NewClient(ctx, os.Getenv("REGISTRY_PROJECT_IDENTIFIER"))
    if err != nil {
        return err
    }

    // Create topic
    var topic *pubsub.Topic
    topic, err = d.pubsubClient.CreateTopic(ctx, topicName)
    if err != nil {
        code := status.Code(err)
        if code == codes.AlreadyExists {
            topic = d.pubsubClient.Topic(topicName)
        } else {
            return err
        }
    }

    // Create subscription
    d.subscription, err = d.pubsubClient.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
            Topic:            topic,
            AckDeadline:      10 * time.Second,
        })
    if err != nil {
        code := status.Code(err)
        if code == codes.AlreadyExists {
            d.subscription = d.pubsubClient.Subscription(subscriptionName)
        }
    }
    log.Printf("%s", d.subscription)

    // Configure subscriber.
    // TODO: Make this configurable.
    d.subscription.ReceiveSettings.MaxOutstandingMessages = 10
    d.subscription.ReceiveSettings.NumGoroutines = runtime.NumCPU()
    // TODO: Create Queue
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

    switch changeType := message.Change; changeType {
        case rpc.Notification_CREATED, rpc.Notification_UPDATED:
            log.Printf("running linter for change type %q", changeType)
        default:
            log.Printf("ignoring change type %q for linting", changeType)
            msg.Ack()
            return
    }

    err := createQueueTask(ctx, message.Resource)
    if err!= nil {
        log.Printf("Error creating queue task: %s", err)
        // Nack message so that it will be retried
        msg.Nack()
        return
    }

    msg.Ack()
    return
}

type workerRequest struct {
    resource string
    command  string
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

    req := &taskspb.CreateTaskRequest{
        Parent: queuePath,
        Task: &taskspb.Task{
            MessageType: &taskspb.Task_HttpRequest{
                HttpRequest: &taskspb.HttpRequest{
                        Url:        workerUrl,
                        HttpMethod: taskspb.HttpMethod_POST,
                },
            },
        },
    }

    wReq := &workerRequest{
        resource: strings.Split(resource, "@")[0],
        command: "registry compute lint",
    }
    jsonReq, _ := json.Marshal(wReq)
    req.Task.GetHttpRequest().Body = []byte(jsonReq)
    log.Print(req)
    createdTask, err := client.CreateTask(ctx, req)
    if err != nil {
            return err
    }

    log.Printf("Created task: %v", createdTask)
    return nil
}