package datastore

import (
	"cloud.google.com/go/datastore"
	"github.com/apigee/registry/server/storage"
)

// Iterator can be used to iterate through results of a query.
type Iterator struct {
	iterator *datastore.Iterator
}

func (it *Iterator) Next(v interface{}) (storage.Key, error) {
	key, err := it.iterator.Next(v)
	if key == nil {
		return nil, err
	}
	return &Key{key: key}, err
}

// GetCursor gets the cursor for the next page of results.
func (it *Iterator) GetCursor(l int) (string, error) {
	if l > 0 {
		nextCursor, err := it.iterator.Cursor()
		if err != nil {
			return "", internalError(err)
		}
		return nextCursor.String(), nil
	}
	return "", nil
}
