// Copyright 2021 Google LLC.
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
	apex "github.com/apex/log"
)

func newApexLogger(opts ...Option) apexAdapter {
	adapter := apexAdapter{
		entry: &apex.Entry{
			Logger: new(apex.Logger),
		},
	}

	for _, opt := range opts {
		opt(&adapter)
	}

	return adapter
}

type apexAdapter struct {
	entry *apex.Entry
}

func (l apexAdapter) Fatal(msg string) {
	l.entry.Fatal(msg)
}

func (l apexAdapter) Fatalf(msg string, v ...interface{}) {
	l.entry.Fatalf(msg, v...)
}

func (l apexAdapter) Error(msg string) {
	l.entry.Error(msg)
}

func (l apexAdapter) Errorf(msg string, v ...interface{}) {
	l.entry.Errorf(msg, v...)
}

func (l apexAdapter) Warn(msg string) {
	l.entry.Warn(msg)
}

func (l apexAdapter) Warnf(msg string, v ...interface{}) {
	l.entry.Warnf(msg, v...)
}

func (l apexAdapter) Info(msg string) {
	l.entry.Info(msg)
}

func (l apexAdapter) Infof(msg string, v ...interface{}) {
	l.entry.Infof(msg, v...)
}

func (l apexAdapter) Debug(msg string) {
	l.entry.Debug(msg)
}

func (l apexAdapter) Debugf(msg string, v ...interface{}) {
	l.entry.Debugf(msg, v...)
}

func (l apexAdapter) WithError(err error) Logger {
	return apexAdapter{
		entry: l.entry.WithError(err),
	}
}

func (l apexAdapter) WithField(k string, v interface{}) Logger {
	return apexAdapter{
		entry: l.entry.WithField(k, v),
	}
}

func (l apexAdapter) WithFields(fields map[string]interface{}) Logger {
	return apexAdapter{
		entry: l.entry.WithFields(apex.Fields(fields)),
	}
}
