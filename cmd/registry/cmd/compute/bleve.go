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
	"sync"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/blevesearch/bleve"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	oas2 "github.com/googleapis/gnostic/openapiv2"
	oas3 "github.com/googleapis/gnostic/openapiv3"
)

const (
	bleveDir = "registry.bleve"
)

var bleveFilter string
var bleveMutex sync.Mutex

func init() {
	computeBleveCmd.Flags().StringVar(&bleveFilter, "filter", "", "Filter option to send with calls")
}

var computeBleveCmd = &cobra.Command{
	Use:   "search-index",
	Short: "Compute a local search index of specs (experimental)",
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
			err = core.ListSpecs(ctx, client, m, bleveFilter, func(spec *rpc.ApiSpec) {
				taskQueue <- &indexSpecTask{
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
		} else {
			log.Fatalf("We don't know how to index %s", name)
		}
	},
}

type indexSpecTask struct {
	ctx      context.Context
	client   connection.Client
	specName string
}

func (task *indexSpecTask) String() string {
	return "index " + task.specName
}

func (task *indexSpecTask) Run() error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(task.ctx, request)
	if err != nil {
		return err
	}
	name := spec.GetName()
	data, err := core.GetBytesForSpec(task.ctx, task.client, spec)
	if err != nil {
		return nil
	}
	var message proto.Message
	if core.IsOpenAPIv2(spec.GetMimeType()) {
		document, err := oas2.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("errors parsing %s", name)
		}
		// remove some fields to simplify the search index
		document.Paths = nil
		document.Definitions = nil
		document.Responses = nil
		document.Parameters = nil
		document.Security = nil
		document.SecurityDefinitions = nil
		message = document
	} else if core.IsOpenAPIv3(spec.GetMimeType()) {
		document, err := oas3.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("errors parsing %s", name)
		}
		// remove some fields to simplify the search index
		document.Paths = nil
		document.Components = nil
		document.Security = nil
		message = document
	} else {
		return fmt.Errorf("unable to generate descriptor for style %s", spec.GetMimeType())
	}

	// The bleve index requires serialized updates.
	bleveMutex.Lock()
	defer bleveMutex.Unlock()
	// Open the index, creating a new one if necessary.
	index, err := bleve.Open(bleveDir)
	if err != nil {
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(bleveDir, mapping)
		if err != nil {
			return err
		}
	}
	defer index.Close()
	// Index the spec.
	log.Printf("indexing %s", task.specName)
	return index.Index(task.specName, message)
}
