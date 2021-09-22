// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graphql

import (
	"time"

	"github.com/graphql-go/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var timestampType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Timestamp",
		Fields: graphql.Fields{
			"seconds": &graphql.Field{
				Type: graphql.Int,
			},
			"nanos": &graphql.Field{
				Type: graphql.Int,
			},
			"rfc3339": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

func representationForTimestamp(timestamp *timestamppb.Timestamp) map[string]interface{} {
	return map[string]interface{}{
		"seconds": timestamp.Seconds,
		"nanos":   timestamp.Nanos,
		"rfc3339": time.Unix(timestamp.Seconds, int64(timestamp.Nanos)).Format(time.RFC3339),
	}
}
