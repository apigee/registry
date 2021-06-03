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

package core

import (
	"context"
	"log"
	"sync"
)

// Task is a generic interface for a runnable operation
type Task interface {
	Run() error
	String() string
}

var wg sync.WaitGroup

func WaitGroup() *sync.WaitGroup {
	return &wg
}

func Worker(ctx context.Context, taskChan <-chan Task) {
	defer wg.Done()
	for task := range taskChan {
		err := task.Run()
		if err != nil {
			log.Printf("%s: %s", task, err)
		}
	}
}
