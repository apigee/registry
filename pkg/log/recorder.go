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
	"sync"

	"github.com/apex/log"
	"github.com/apex/log/handlers/multi"
)

// Creates a Logger with a Recorder for accessing created log entries
func NewWithRecorder(opts ...Option) (Logger, *recorder) {
	memory := &recorder{}
	multiOption := func(l *apexAdapter) {
		l.entry.Logger.Handler = multi.New(memory, l.entry.Logger.Handler)
	}
	opts = append(opts, multiOption)
	return NewLogger(opts...), memory
}

type entry struct {
	le *log.Entry
}

func (e *entry) Message() string {
	if e == nil {
		return ""
	}
	return e.le.Message
}

type recorder struct {
	sync.Mutex
	LogEntries []*log.Entry
}

func (m *recorder) Entries() []*entry {
	m.Lock()
	defer m.Unlock()
	entries := []*entry{}
	for _, e := range m.LogEntries {
		entries = append(entries, &entry{e})
	}
	return entries
}

func (m *recorder) LastEntry() *entry {
	m.Lock()
	defer m.Unlock()
	if len(m.LogEntries) == 0 {
		return nil
	}
	return &entry{m.LogEntries[len(m.LogEntries)-1]}
}

func (m *recorder) ClearEntries() {
	m.Lock()
	defer m.Unlock()
	m.LogEntries = nil
}

// HandleLog implements log.Handler.
func (h *recorder) HandleLog(e *log.Entry) error {
	h.Lock()
	defer h.Unlock()
	h.LogEntries = append(h.LogEntries, e)
	return nil
}
