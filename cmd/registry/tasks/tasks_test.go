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
	"testing"
)

func TestWorkerPool(t *testing.T) {
	ctx := context.Background()
	jobs := 100

	taskQueue, wait := WorkerPool(ctx, jobs)
	defer wait()

	for i := 0; i < 1000; i++ {
		taskQueue <- &doNothingTask{i: i}
	}
}

type doNothingTask struct {
	i int
}

func (task *doNothingTask) String() string {
	return fmt.Sprintf("do nothing %d", task.i)
}

func (task *doNothingTask) Run(ctx context.Context) error {
	return nil
}

func TestWorkerPoolWithWarnings(t *testing.T) {
	ctx := context.Background()
	jobs := 1

	taskQueue, wait := WorkerPoolWithWarnings(ctx, jobs)
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
