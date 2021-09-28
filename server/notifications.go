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

	"cloud.google.com/go/pubsub"
	"github.com/apex/log"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

	if _, err := client.CreateTopic(ctx, TopicName); err != nil && status.Code(err) != codes.AlreadyExists {
		return err
	}

	topic := client.Topic(TopicName)
	defer topic.Stop()

	msg, err := protojson.Marshal(&rpc.Notification{
		Change:     change,
		Resource:   resource,
		ChangeTime: timestamppb.Now(),
	})
	if err != nil {
		return err
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	id, err := result.Get(ctx)
	if err != nil {
		return err
	}

	notificationTotal++
	log.Infof("^^ [%03d] %+s", notificationTotal, msg)
	log.Infof("Published a message with a message ID: %s", id)

	return nil
}
