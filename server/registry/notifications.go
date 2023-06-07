// Copyright 2020 Google LLC.
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
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/rpc"
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

	client, err := s.getPubSubClient(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get PubSub client.")
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
