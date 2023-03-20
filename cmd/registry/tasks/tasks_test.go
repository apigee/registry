// Copyright 2023 Google LLC. All Rights Reserved.
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
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	ctx := context.Background()
	jobs := 100
	counter := new(atomic.Int32)

	taskQueue, wait := WorkerPool(ctx, jobs)
	defer wait()

	for i := 0; i < 1000; i++ {
		taskQueue <- &incrTask{counter}
	}

	count := counter.Load()
	if count != int32(1000) {
		t.Errorf("expected %d got: %d", 1000, count)
	}
}

type incrTask struct {
	counter *atomic.Int32
}

func (t *incrTask) String() string {
	return "add 1"
}

func (t *incrTask) Run(ctx context.Context) error {
	t.counter.Add(1)
	time.Sleep(time.Millisecond) // make the task last a moment
	return nil
}

func TestWorkerPoolWithWarnings(t *testing.T) {
	ctx := context.Background()
	jobs := 1

	taskQueue, wait := WorkerPool(ctx, jobs)
	defer wait()

	for i := 0; i < 10; i++ {
		taskQueue <- &failTask{i: i}
	}
}

type failTask struct {
	i int
}

func (task *failTask) String() string {
	return fmt.Sprintf("do nothing %d", task.i)
}

func (task *failTask) Run(ctx context.Context) error {
	return errors.New("fail")
}
