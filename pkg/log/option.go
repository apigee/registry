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
	"io"

	apex "github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
)

// Options configure the logger.
type Option func(*apexAdapter)

// Configures the logger to print logs with a minimum severity level.
var (
	DebugLevel Option = func(l *apexAdapter) {
		l.entry.Logger.Level = apex.DebugLevel
	}

	InfoLevel Option = func(l *apexAdapter) {
		l.entry.Logger.Level = apex.InfoLevel
	}

	WarnLevel Option = func(l *apexAdapter) {
		l.entry.Logger.Level = apex.WarnLevel
	}

	ErrorLevel Option = func(l *apexAdapter) {
		l.entry.Logger.Level = apex.ErrorLevel
	}

	FatalLevel Option = func(l *apexAdapter) {
		l.entry.Logger.Level = apex.FatalLevel
	}
)

// JSONFormat configures the logger to print logs as JSON.
func JSONFormat(w io.Writer) Option {
	return func(l *apexAdapter) {
		l.entry.Logger.Handler = json.New(w)
	}
}

// TextFormat configures the logger to print logs as human-friendly text.
func TextFormat(w io.Writer) Option {
	return func(l *apexAdapter) {
		l.entry.Logger.Handler = text.New(w)
	}
}
