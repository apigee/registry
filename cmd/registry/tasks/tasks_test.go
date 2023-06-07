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
	"errors"
	"sync"
	"testing"
)

func TestWorkerPoolContinueOnError(t *testing.T) {
	ctx := context.Background()
	jobs := 100
	counter := new(atomicInt32)
	total := 1000

	func() {
		taskQueue, wait := WorkerPool(ctx, jobs, true)
		for i := 0; i < total; i++ {
			taskQueue <- &incrTask{counter}
		}
		if err := wait(); err != nil {
			t.Error(err)
		}
	}()

	got := counter.Load()
	if got != int32(total) {
		t.Errorf("want %d got: %d", total, got)
	}
}

func TestWorkerPoolStopOnError(t *testing.T) {
	ctx := context.Background()
	jobs := 100
	counter := new(atomicInt32)
	total := 1000
	errorAt := 500

	func() {
		taskQueue, wait := WorkerPool(ctx, jobs, false)
		for i := 0; i < total; i++ {
			if i == errorAt {
				taskQueue <- &failTask{}
			}
			taskQueue <- &incrTask{counter}
		}
		if err := wait(); err == nil {
			t.Log("expected error")
		}
	}()

	got := counter.Load()
	if got < int32(errorAt) || got > int32(total) {
		t.Errorf("want from %d to %d, got: %d", errorAt, total, got)
	}
}

type incrTask struct {
	counter *atomicInt32
}

func (t *incrTask) String() string {
	return "add 1"
}

func (t *incrTask) Run(ctx context.Context) error {
	t.counter.Add(1)
	return nil
}

func TestWorkerPoolIgnoreError(t *testing.T) {
	ctx := context.Background()
	jobs := 1

	taskQueue, wait := WorkerPoolIgnoreError(ctx, jobs)
	defer wait()
	for i := 0; i < 2; i++ {
		taskQueue <- &failTask{}
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

// can be replaced by atomic.Int32 when we move to go 1.19+
type atomicInt32 struct {
	sync.RWMutex
	v int32
}

// Load atomically loads and returns the value stored in x.
func (a *atomicInt32) Load() int32 {
	a.RLock()
	defer a.RUnlock()
	return a.v
}

// Add atomically adds delta to x and returns the new value.
func (a *atomicInt32) Add(delta int32) int32 {
	a.Lock()
	defer a.Unlock()
	a.v = a.v + delta
	return a.v
}
