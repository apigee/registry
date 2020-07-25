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
