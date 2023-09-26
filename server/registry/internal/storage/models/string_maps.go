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

package models

import (
	"encoding/json"

	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

func bytesForMap(entries map[string]string) ([]byte, error) {
	return proto.Marshal(&rpc.Map{Entries: entries})
}

func mapForBytes(b []byte) (map[string]string, error) {
	m := &rpc.Map{}
	if err := proto.Unmarshal(b, m); err != nil {
		return nil, err
	}
	return m.Entries, nil
}

func jbytesForMap(entries map[string]string) ([]byte, error) {
	if entries == nil {
		entries = map[string]string{}
	}
	return json.Marshal(entries)
}

func jmapForBytes(b []byte) (map[string]string, error) {
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
