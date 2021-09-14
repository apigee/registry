// Copyright 2020 Google LLC. All Rights Reserved.
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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/apigee/registry/cmd/registry/cmd"
	"github.com/google/uuid"
)

// Initialize global default logger with unique process identifier.
func init() {
	logger := &log.Logger{
		Level:   log.DebugLevel,
		Handler: text.Default,
	}
	log.Log = logger.WithField("uid", fmt.Sprintf("[ %.8s ] ", uuid.New()))
}

func main() {
	ctx := context.Background()
	cmd := cmd.Command(ctx)
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
