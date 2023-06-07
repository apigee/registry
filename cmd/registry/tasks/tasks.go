// Copyright 2023 Google LLC.
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

// WorkerPool creates a waitgroup and a taskQueue for a worker pool.
// It will create "n" workers which will listen for Tasks on the taskQueue.
// The return value is the taskQueue and a wait function.
// Do not directly close the taskQueue, use the wait function.
// Clients should add new tasks to this taskQueue and call the wait function
// when done. If a worker fails with an error and continueOnError is false,
// the task context will be canceled, instructing the queue and workers to
// terminate. Tasks should check ctx as appropriate. The first error encountered
// will be returned from the wait function and all errors are logged at Warn level.
func WorkerPool(ctx context.Context, n int, continueOnError bool) (chan<- Task, func() error) {
	var eg *errgroup.Group
	if continueOnError {
		eg = &errgroup.Group{}
	} else {
		eg, ctx = errgroup.WithContext(ctx)
	}
	eg.SetLimit(n) // pool size
	taskQueue := make(chan Task, 1024)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for task := range taskQueue {
			select {
			case <-ctx.Done():
				return
			default:
				t := task
				f := func() error {
					err := t.Run(ctx)
					if err != nil {
						log.FromContext(ctx).WithError(err).Warnf("task failed: %s", t)
					}
					return err
				}
				eg.Go(f) // blocks at n runners
			}
		}
	}()

	wait := func() error {
		close(taskQueue)
		wg.Wait()        // wait for all work to be added
		err := eg.Wait() // wait for all work to be completed
		if err != nil {
			log.FromContext(ctx).WithError(err).Warnf("WorkerPool terminated")
		}
		return err
	}

	return taskQueue, wait
}

// WorkerPoolIgnoreError instantiates a WorkerPool with continueOnError=true and
// returns a wait function that captures and logs the error at Error level. This
// is convenient for defer.
func WorkerPoolIgnoreError(ctx context.Context, n int) (chan<- Task, func()) {
	wp, w := WorkerPool(ctx, n, true)
	wait := func() {
		if err := w(); err != nil {
			log.FromContext(ctx).WithError(err).Errorf("unhandled WorkerPool error")
		}
	}
	return wp, wait
}
