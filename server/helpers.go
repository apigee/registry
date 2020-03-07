package server

import "cloud.google.com/go/datastore"

func queryApplyPageSize(q *datastore.Query, pageSize int32) *datastore.Query {
	if pageSize > 1000 {
		return q.Limit(1000)
	}
	if pageSize <= 0 {
		return q.Limit(50)
	}
	return q.Limit(int(pageSize))
}

func queryApplyCursor(q *datastore.Query, cursorStr string) (*datastore.Query, error) {
	if cursorStr != "" {
		cursor, err := datastore.DecodeCursor(cursorStr)
		if err != nil {
			return nil, internalError(err)
		}
		q = q.Start(cursor)
	}
	return q, nil
}

// Get the cursor for the next page of results.
func iteratorGetCursor(it *datastore.Iterator, l int) (string, error) {
	if l > 0 {
		nextCursor, err := it.Cursor()
		if err != nil {
			return "", internalError(err)
		}
		return nextCursor.String(), nil
	}
	return "", nil
}
