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

package models

import (
	"sort"

	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

func bytesForMap(m map[string]string) ([]byte, error) {
	pairs := make([]*rpc.OrderedPair, 0)
	for k, v := range m {
		pairs = append(pairs, &rpc.OrderedPair{Name: k, Value: v})
	}
	sort.Slice(pairs,
		func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
	return proto.Marshal(&rpc.OrderedMap{Pairs: pairs})
}

func mapForBytes(b []byte) (map[string]string, error) {
	om := &rpc.OrderedMap{}
	if err := proto.Unmarshal(b, om); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, p := range om.Pairs {
		m[p.Name] = p.Value
	}
	return m, nil
}
