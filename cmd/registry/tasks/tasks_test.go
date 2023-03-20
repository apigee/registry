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
	"sync/atomic"
	"testing"
)

func TestWorkerPool(t *testing.T) {
	ctx := context.Background()
	jobs := 100
	counter := new(atomic.Int32)
	want := 1000

	func() {
		taskQueue, wait := WorkerPool(ctx, jobs)
		for i := 0; i < want; i++ {
			taskQueue <- &incrTask{counter}
		}
		if err := wait(); err != nil {
			t.Error(err)
		}
	}()

	got := counter.Load()
	if got != int32(want) {
		t.Errorf("want %d got: %d", want, got)
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
	return nil
}

func TestWorkerPoolWithWarnings(t *testing.T) {
	ctx := context.Background()
	jobs := 1

	taskQueue, wait := WorkerPool(ctx, jobs)
	for i := 0; i < 2; i++ {
		taskQueue <- &failTask{}
	}
	if err := wait(); err == nil {
		t.Error("want error, got nil")
	}
}

type failTask struct {
}

func (task *failTask) String() string {
	return "fail task"
}

func (task *failTask) Run(ctx context.Context) error {
	return errors.New("fail")
}
