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
	"context"
	"os"
)

// Logger describes the interface of a structured logger.
type Logger interface {
	Fatal(string)
	Fatalf(string, ...interface{})

	Error(string)
	Errorf(string, ...interface{})

	Warn(string)
	Warnf(string, ...interface{})

	Info(string)
	Infof(string, ...interface{})

	Debug(string)
	Debugf(string, ...interface{})

	WithError(error) Logger
	WithField(string, interface{}) Logger
	WithFields(map[string]interface{}) Logger
}

var defaultOptions = []Option{
	TextFormat(os.Stderr),
	InfoLevel,
}

// NewLogger returns a logger configured with the provided options.
func NewLogger(opts ...Option) Logger {
	opts = append(defaultOptions, opts...) // Override defaults with user options.
	return newApexLogger(opts...)
}

func Fatal(ctx context.Context, msg string) {
	FromContext(ctx).Fatal(msg)
}

func Fatalf(ctx context.Context, msg string, v ...interface{}) {
	FromContext(ctx).Fatalf(msg, v...)
}

func Error(ctx context.Context, msg string) {
	FromContext(ctx).Error(msg)
}

func Errorf(ctx context.Context, msg string, v ...interface{}) {
	FromContext(ctx).Errorf(msg, v...)
}

func Warn(ctx context.Context, msg string) {
	FromContext(ctx).Warn(msg)
}

func Warnf(ctx context.Context, msg string, v ...interface{}) {
	FromContext(ctx).Warnf(msg, v...)
}

func Info(ctx context.Context, msg string) {
	FromContext(ctx).Info(msg)
}

func Infof(ctx context.Context, msg string, v ...interface{}) {
	FromContext(ctx).Infof(msg, v...)
}

func Debug(ctx context.Context, msg string) {
	FromContext(ctx).Debug(msg)
}

func Debugf(ctx context.Context, msg string, v ...interface{}) {
	FromContext(ctx).Debugf(msg, v...)
}
