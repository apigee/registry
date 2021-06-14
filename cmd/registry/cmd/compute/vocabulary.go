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

package compute

import (
	"context"
	"fmt"
	"log"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/googleapis/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	discovery "github.com/googleapis/gnostic/discovery"
	metrics "github.com/googleapis/gnostic/metrics"
	oas2 "github.com/googleapis/gnostic/openapiv2"
	oas3 "github.com/googleapis/gnostic/openapiv3"
)

var computeVocabularyCmd = &cobra.Command{
	Use:   "vocabulary",
	Short: "Compute vocabularies of API specs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		// Initialize task queue.
		taskQueue := make(chan core.Task, 1024)
		workerCount := 64
		for i := 0; i < workerCount; i++ {
			core.WaitGroup().Add(1)
			go core.Worker(ctx, taskQueue)
		}
		// Generate tasks.
		name := args[0]
		if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
			// Iterate through a collection of specs and summarize each.
			err = core.ListSpecs(ctx, client, m, computeFilter, func(spec *rpc.ApiSpec) {
				taskQueue <- &computeVocabularyTask{
					ctx:      ctx,
					client:   client,
					specName: spec.Name,
				}
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			close(taskQueue)
			core.WaitGroup().Wait()
		}
	},
}

type computeVocabularyTask struct {
	ctx      context.Context
	client   connection.Client
	specName string
}

func (task *computeVocabularyTask) String() string {
	return "compute vocabulary " + task.specName
}

func (task *computeVocabularyTask) Run() error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(task.ctx, request)
	if err != nil {
		return err
	}
	relation := "vocabulary"
	log.Printf("computing %s/artifacts/%s", spec.Name, relation)
	var vocab *metrics.Vocabulary
	if core.IsOpenAPIv2(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(task.ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := oas2.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI: %s", spec.Name)
		}
		vocab = vocabulary.NewVocabularyFromOpenAPIv2(document)
	} else if core.IsOpenAPIv3(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(task.ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := oas3.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI: %s", spec.Name)
		}
		vocab = vocabulary.NewVocabularyFromOpenAPIv3(document)
	} else if core.IsDiscovery(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(task.ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := discovery.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("invalid Discovery: %s", spec.Name)
		}
		vocab = vocabulary.NewVocabularyFromDiscovery(document)
	} else if core.IsProto(spec.GetMimeType()) && core.IsZipArchive(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(task.ctx, task.client, spec)
		if err != nil {
			return nil
		}
		vocab, err = core.NewVocabularyFromZippedProtos(data)
		if err != nil {
			return fmt.Errorf("error processing protos: %s", spec.Name)
		}
	} else {
		return fmt.Errorf("we don't know how to summarize %s", spec.Name)
	}
	subject := spec.GetName()
	messageData, _ := proto.Marshal(vocab)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("gnostic.metrics.Vocabulary"),
		Contents: messageData,
	}
	err = core.SetArtifact(task.ctx, task.client, artifact)
	if err != nil {
		return err
	}
	return nil
}
