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

package registry

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const TopicName = "registry-events"

func (s *RegistryServer) notify(ctx context.Context, change rpc.Notification_Change, resource string) {
	if !s.notifyEnabled {
		return
	}

	logger := log.FromContext(ctx)
	if s.projectID == "" {
		logger.Warn("Notifications are enabled but project ID is not set. Skipping notification.")
		return
	}

	client, err := pubsub.NewClient(ctx, s.projectID)
	if err != nil {
		logger.WithError(err).Error("Failed to create PubSub client.")
		return
	}
	defer client.Close()

	if _, err := client.CreateTopic(ctx, TopicName); err != nil && status.Code(err) != codes.AlreadyExists {
		logger.WithError(err).Error("Failed to create PubSub topic.")
		return
	}

	notification := &rpc.Notification{
		Change:     change,
		Resource:   resource,
		ChangeTime: timestamppb.Now(),
	}

	msg, err := protojson.Marshal(notification)
	if err != nil {
		logger.WithError(err).Errorf("Failed to serialize notification: %v", notification)
		return
	}

	topic := client.Topic(TopicName)
	defer topic.Stop()

	result := topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	id, err := result.Get(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to publish notification.")
		return
	}

	logger.Infof("Published notification with message ID: %s", id)
}
