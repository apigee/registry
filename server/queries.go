package server

import "cloud.google.com/go/datastore"

func boundPageSize(pageSize int32) int {
	if pageSize > 1000 {
		return 1000
	}
	if pageSize <= 0 {
		return 50
	}
	return int(pageSize)
}

func queryApplyPageSize(q *datastore.Query, pageSize int32) *datastore.Query {
	return q.Limit(boundPageSize(pageSize))
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
