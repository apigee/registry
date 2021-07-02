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

package server

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/apigee/registry/rpc"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const verbose = false

const TopicName = "registry-events"

var notificationTotal int

func (s *RegistryServer) notify(ctx context.Context, change rpc.Notification_Change, resource string) error {
	if !s.notifyEnabled || s.projectID == "" {
		return nil
	}
	client, err := pubsub.NewClient(ctx, s.projectID)
	if err != nil {
		return err
	}
	defer client.Close()
	// Ensure that topic exists.
	{
		_, err := client.CreateTopic(ctx, TopicName)
		if err != nil {
			code := status.Code(err)
			if code != codes.AlreadyExists {
				return err
			}
		}
	}
	// Get the topic.
	topic := client.Topic(TopicName)
	defer topic.Stop()
	// Create the notification
	n := &rpc.Notification{}
	n.Change = change
	n.Resource = resource
	n.ChangeTime, err = ptypes.TimestampProto(time.Now())
	// Marshal the notification.
	m, err := (&jsonpb.Marshaler{}).MarshalToString(n)
	if err != nil {
		return err
	}
	// Send the notification.
	notificationTotal++
	log.Printf("^^ [%03d] %+s", notificationTotal, m)
	var results []*pubsub.PublishResult
	r := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(m),
	})
	results = append(results, r)
	for _, r := range results {
		id, err := r.Get(ctx)
		if err != nil {
			return err
		}
		if verbose {
			log.Printf("Published a message with a message ID: %s", id)
		}
	}
	return nil
}
