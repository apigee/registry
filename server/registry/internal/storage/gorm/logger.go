// Copyright 2021 Google LLC. All Rights Reserved.
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

package gorm

import (
	"context"
	"time"

	"github.com/apex/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type gormLogger struct {
	Logger        log.Interface
	SlowThreshold time.Duration
}

func NewGormLogger(ctx context.Context) logger.Interface {
	return gormLogger{
		Logger:        log.FromContext(ctx),
		SlowThreshold: 100 * time.Millisecond,
	}
}

func (l gormLogger) LogMode(logger.LogLevel) logger.Interface {
	return l // Ignore gorm's log levels. Our logger handles levels independently.
}

func (l gormLogger) Info(context.Context, string, ...interface{}) {
	// Ignore non-trace calls. We don't know what calls this.
}

func (l gormLogger) Warn(context.Context, string, ...interface{}) {
	// Ignore non-trace calls. We don't know what calls this.
}

func (l gormLogger) Error(context.Context, string, ...interface{}) {
	// Ignore non-trace calls. We don't know what calls this.
}

func (l gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, _ := fc()
	entry := l.Logger.WithFields(log.Fields{
		"query":    sql,
		"duration": time.Since(begin),
	})

	if err != nil && err != gorm.ErrRecordNotFound {
		entry.WithError(err).Error("Failed database operation.")
	} else if time.Since(begin) > l.SlowThreshold {
		entry.Warn("Slow database operation.")
	} else {
		entry.Info("Database operation.")
	}
}
