// Copyright 2022 Google LLC. All Rights Reserved.
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
	"fmt"
	"testing"

	"cloud.google.com/go/pubsub/pstest"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"
)

func TestNotifications(t *testing.T) {
	ctx := context.Background()

	pubSubTest := pstest.NewServer()
	defer pubSubTest.Close()

	// Connect to the server without using TLS
	conn, err := grpc.Dial(pubSubTest.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	projectID := "myproject"
	server, err := New(Config{
		Database:         "sqlite3",
		DBConfig:         fmt.Sprintf("%s/registry.db", t.TempDir()),
		APIClientOptions: []option.ClientOption{option.WithGRPCConn(conn)},
		ProjectID:        projectID,
		Notify:           true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	server.notify(ctx, rpc.Notification_CREATED, "resource")
	pubSubTest.Wait()

	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, TopicName)
	topic, err := pubSubTest.GServer.GetTopic(ctx, &pubsub.GetTopicRequest{Topic: topicName})
	if err != nil {
		t.Fatal(err)
	}
	if topic == nil {
		t.Errorf("Topic %q not found", TopicName)
	}

	ms := pubSubTest.Messages()
	if len(ms) != 1 {
		t.Errorf("Expected 1 message, got %d", len(ms))
	}
}

func TestNotificationErrors(t *testing.T) {
	logger, rec := log.NewWithRecorder()
	ctx := log.NewContext(context.Background(), logger)

	server := RegistryServer{
		notifyEnabled: true,
	}
	server.notify(ctx, rpc.Notification_CREATED, "resource")

	entry := rec.LastEntry()
	want := "Notifications are enabled but project ID is not set. Skipping notification."
	if want != entry.Message() {
		t.Errorf(cmp.Diff(want, entry.Message))
	}

	server.projectID = "id"
	server.notify(ctx, rpc.Notification_CREATED, "resource")
	entry = rec.LastEntry()
	want = "Failed to get PubSub client."
	if want != entry.Message() {
		t.Errorf(cmp.Diff(want, entry.Message()))
	}
}
