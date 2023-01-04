// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lint

import (
	"time"
)

func NewResponse() *Response {
	r := &Response{
		RunTime:  time.Now(),
		Problems: make([]Problem, 0),
		receiver: make(chan checked),
	}
	go func() {
		for res := range r.receiver {
			r.Problems = append(r.Problems, res.problems...)
			if res.err != nil {
				r.Error = res.err
			}
		}
	}()

	return r
}

type checked struct {
	resource Resource
	problems []Problem
	err      error // populated if panic
}

// Response collects the results of running the Rules.
type Response struct {
	RunTime  time.Time `json:"time" yaml:"time"`
	Problems []Problem `json:"problems" yaml:"problems"`
	Error    error     `json:"error,omitempty" yaml:"error,omitempty"` // populated if panic
	receiver chan checked
}

func (r *Response) checked(res Resource, probs []Problem, err error) {
	r.receiver <- checked{
		resource: res,
		problems: probs,
		err:      err,
	}
}
