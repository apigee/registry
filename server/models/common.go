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

package models

import (
	"context"
	"log"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

const verbose = false

// DeleteAllMatches deletes all entities matching a query.
func DeleteAllMatches(ctx context.Context, client *datastore.Client, q *datastore.Query) error {
	it := client.Run(ctx, q.Distinct())
	key, err := it.Next(nil)
	keys := make([]*datastore.Key, 0)
	for err == nil {
		keys = append(keys, key)
		key, err = it.Next(nil)
		if len(keys) == 500 {
			if verbose {
				log.Printf("Deleting %d %s entities", len(keys), keys[0].Kind)
			}
			err = client.DeleteMulti(ctx, keys)
			if err != nil {
				return err
			}
			keys = make([]*datastore.Key, 0)
		}
	}
	if err != iterator.Done {
		return err
	}
	if len(keys) > 0 {
		if verbose {
			log.Printf("Deleting %d %s entities", len(keys), keys[0].Kind)
		}
		return client.DeleteMulti(ctx, keys)
	}
	return nil
}
