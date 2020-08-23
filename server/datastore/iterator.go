package datastore

import "cloud.google.com/go/datastore"

// Iterator can be used to iterate through results of a query.
type Iterator = datastore.Iterator

// IteratorGetCursor gets the cursor for the next page of results.
func (client *Client) IteratorGetCursor(it *Iterator, l int) (string, error) {
	if l > 0 {
		nextCursor, err := it.Cursor()
		if err != nil {
			return "", internalError(err)
		}
		return nextCursor.String(), nil
	}
	return "", nil
}
