// Copyright 2022 Google LLC.
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

package log

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRecorder(t *testing.T) {
	tests := []struct {
		desc            string
		level           Option
		wantNumEntries  int
		wantLastMessage string
	}{
		{
			"error",
			ErrorLevel,
			2,
			"test error",
		},
		{
			"warn",
			WarnLevel,
			4,
			"test warn",
		},
		{
			"info",
			InfoLevel,
			6,
			"test info",
		},
		{
			"debug",
			DebugLevel,
			8,
			"test debug",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			logger, rec := NewWithRecorder(test.level)
			ctx := NewContext(context.Background(), logger)

			Error(ctx, "test error")
			Errorf(ctx, "test error")
			Warn(ctx, "test warn")
			Warnf(ctx, "test warn")
			Info(ctx, "test info")
			Infof(ctx, "test info")
			Debug(ctx, "test debug")
			Debugf(ctx, "test debug")

			if len(rec.Entries()) != test.wantNumEntries {
				t.Errorf("want %d entries, got: %d", test.wantNumEntries, len(rec.Entries()))
			}

			lastEntry := rec.LastEntry()
			if test.wantLastMessage != lastEntry.Message() {
				t.Errorf(cmp.Diff(test.wantLastMessage, lastEntry.Message()))
			}
		})
	}
}
