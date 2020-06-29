package models

import (
	"context"
	"fmt"
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

func deleteAllMatches(ctx context.Context, client *datastore.Client, q *datastore.Query) error {
	it := client.Run(ctx, q.Distinct())
	key, err := it.Next(nil)
	keys := make([]*datastore.Key, 0)
	for err == nil {
		keys = append(keys, key)
		key, err = it.Next(nil)
	}
	if err != iterator.Done {
		return err
	}
	return client.DeleteMulti(ctx, keys)
}
