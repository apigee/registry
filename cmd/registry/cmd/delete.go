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

package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes matching entities and their children.",
	Long:  "Deletes matching entities and their children.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("delete called with %+v\n", args)
		name := args[0]
		if m := names.ApisRegexp().FindAllStringSubmatch(name, -1); m != nil {
			deleteAllApisInProject(m[0][1])
		} else if m := names.PropertiesRegexp().FindAllStringSubmatch(name, -1); m != nil {
			deleteAllPropertiesInProject(m[0][1])
		} else if m := names.LabelsRegexp().FindAllStringSubmatch(name, -1); m != nil {
			deleteAllLabelsInProject(m[0][1])
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func deleteAllApisInProject(projectID string) {
	ctx := context.TODO()
	client, err := connection.NewClient(ctx)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	request := &rpc.ListApisRequest{
		Parent: "projects/" + projectID,
	}
	log.Printf("%+v", request)
	it := client.ListApis(ctx, request)
	names := make([]string, 0)
	for {
		api, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalf("%s", err.Error())
		}
		names = append(names, api.Name)
	}
	log.Printf("%+v", names)
	count := len(names)
	completions := make(chan int)
	for _, name := range names {
		go func(name string) {
			request := &rpc.DeleteApiRequest{}
			request.Name = name
			err = client.DeleteApi(ctx, request)
			completions <- 1
		}(name)
	}
	for i := 0; i < count; i++ {
		<-completions
		fmt.Printf("COMPLETE: %d\n", i+1)
	}
}

func deleteAllPropertiesInProject(projectID string) {
	ctx := context.TODO()
	client, err := connection.NewClient(ctx)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	request := &rpc.ListPropertiesRequest{
		Parent: "projects/" + projectID,
	}
	log.Printf("%+v", request)
	it := client.ListProperties(ctx, request)
	names := make([]string, 0)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalf("%s", err.Error())
		}
		names = append(names, property.Name)
	}
	log.Printf("%+v", names)
	count := len(names)
	completions := make(chan int)
	for _, name := range names {
		go func(name string) {
			request := &rpc.DeletePropertyRequest{}
			request.Name = name
			err = client.DeleteProperty(ctx, request)
			completions <- 1
		}(name)
	}
	for i := 0; i < count; i++ {
		<-completions
		fmt.Printf("COMPLETE: %d\n", i+1)
	}
}

func deleteAllLabelsInProject(projectID string) {
	ctx := context.TODO()
	client, err := connection.NewClient(ctx)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	request := &rpc.ListLabelsRequest{
		Parent: "projects/" + projectID,
	}
	log.Printf("%+v", request)
	it := client.ListLabels(ctx, request)
	names := make([]string, 0)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalf("%s", err.Error())
		}
		names = append(names, property.Name)
	}
	log.Printf("%+v", names)
	count := len(names)
	completions := make(chan int)
	for _, name := range names {
		go func(name string) {
			request := &rpc.DeleteLabelRequest{}
			request.Name = name
			err = client.DeleteLabel(ctx, request)
			completions <- 1
		}(name)
	}
	for i := 0; i < count; i++ {
		<-completions
		fmt.Printf("COMPLETE: %d\n", i+1)
	}
}
