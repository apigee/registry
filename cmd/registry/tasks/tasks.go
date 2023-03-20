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

package tasks

import (
	"context"
	"sync"

	"github.com/apigee/registry/pkg/log"
	"golang.org/x/sync/errgroup"
)

// Task is a generic interface for a runnable operation
type Task interface {
	Run(ctx context.Context) error
	String() string
}

// This function creates a waitgroup and a taskQueue for the workerPool.
// It will create "n" workers which will listen for Tasks on the taskQueue.
// It returns the taskQueue and a wait func.
// The clients should add new tasks to this taskQueue
// and call the call the wait func when done.
// Do not separately close the taskQueue, make use of the wait func.
func WorkerPool(ctx context.Context, n int) (chan<- Task, func() error) {
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(n) // pool size
	taskQueue := make(chan Task, 1024)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for task := range taskQueue {
			select {
			case <-ctx.Done():
				return
			default:
				t := task
				f := func() error {
					err := t.Run(ctx)
					if err != nil {
						log.FromContext(ctx).WithError(err).Warnf("Task failed: %s", task)
					}
					return err
				}
				eg.Go(f) // blocks at n runners
			}
		}
		wg.Done()
	}()

	wait := func() error {
		close(taskQueue)
		wg.Wait()        // wait for all work to be added
		err := eg.Wait() // wait for all work to be completed
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("WorkerPool closed on err: %v", err)
		}
		return err
	}

	return taskQueue, wait
}
