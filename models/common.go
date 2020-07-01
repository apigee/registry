package models

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

// We might extend this to all characters that do not require escaping.
// See "Resource ID Segments" in https://aip.dev/122.
const nameRegex = "([a-zA-Z0-9-_\\.]+)"

// Generated revision names are lowercase hex strings, but we also
// allow user-specified revision tags which can be mixed-case strings
// containing dashes.
const revisionRegex = "(@[a-zA-z0-9-]+)?"

func validateID(id string) error {
	r := regexp.MustCompile("^" + nameRegex + "$")
	m := r.FindAllStringSubmatch(id, -1)
	if m == nil {
		return fmt.Errorf("invalid id '%s'", id)
	}
	return nil
}

func validateRevision(s string) error {
	r := regexp.MustCompile("^" + revisionRegex + "$")
	m := r.FindAllStringSubmatch(s, -1)
	if m == nil {
		return fmt.Errorf("invalid revision '%s'", s)
	}
	return nil
}

// DeleteAllMatches deletes all entities matching a query.
func DeleteAllMatches(ctx context.Context, client *datastore.Client, q *datastore.Query) error {
	it := client.Run(ctx, q.Distinct())
	key, err := it.Next(nil)
	keys := make([]*datastore.Key, 0)
	for err == nil {
		keys = append(keys, key)
		key, err = it.Next(nil)
		if len(keys) == 500 {
			log.Printf("Deleting %d %s entities", len(keys), keys[0].Kind)
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
		log.Printf("Deleting %d %s entities", len(keys), keys[0].Kind)
		return client.DeleteMulti(ctx, keys)
	}
	return nil
}
