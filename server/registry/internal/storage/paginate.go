// Copyright 2022 Google LLC.
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

package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
)

// PageOptions contains custom arguments for listing requests.
type PageOptions struct {
	// Size is the maximum number of resources to include in the response.
	// If unspecified, page size defaults to 50. Values above 1000 are coerced to 1000.
	Size int32
	// Filter is the filter string for this listing request, as described at https://google.aip.dev/160.
	Filter string
	// Order is the sorting order for this listing request, as described at https://google.aip.dev/132#ordering.
	Order string
	// Token is a value returned from with a previous page in a series of listing requests.
	// If specified, listing will continue from the end of the previous page. Otherwise,
	// the first page in a listing series will be returned.
	Token string
}

// token contains information to share between sequential page iterators.
type token struct {
	// Offset is the number of resources that should be skipped before the page begins.
	// It should be set to the number of resources already returned.
	Offset int
	// Filter is the filter string for this listing request. It should be consistent between sequential pages.
	Filter string
	// Order is the sorting order for this listing request. It should be consistent between sequential pages.
	Order string
}

// ValidateFilter returns an error if the new filter doesn't match the token's encoded filter.
// When the token represents the first page, any filter is valid and no error will be returned.
func (t token) ValidateFilter(newFilter string) error {
	if t.Offset > 0 && newFilter != t.Filter {
		return fmt.Errorf("new filter does not match previous filter %q", t.Filter)
	}

	return nil
}

// ValidateOrder returns an error if the new order doesn't match the token's encoded order, or
// if the format of the ordering string is invalid.
// When the token represents the first page, any order is valid and no error will be returned.
func (t token) ValidateOrder(newOrder string) error {
	if t.Offset > 0 && newOrder != t.Order {
		return fmt.Errorf("new order does not match previous order %q", t.Order)
	}

	return nil
}

// encodeToken converts a token struct into an opaque string that can be converted back into struct form using decodeToken().
func encodeToken(o token) (string, error) {
	var encoding bytes.Buffer

	encoder := gob.NewEncoder(&encoding)
	if err := encoder.Encode(o); err != nil {
		return "", fmt.Errorf("failed to encode token: %s", err)
	}

	return base64.StdEncoding.EncodeToString(encoding.Bytes()), nil
}

// decodeToken converts a string returned from encodeToken() back into an equivalent token struct.
// Empty encoding strings are decoded without error to a zero-value token struct.
func decodeToken(encoded string) (token, error) {
	if encoded == "" {
		return token{}, nil
	}

	decoding, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return token{}, fmt.Errorf("failed to decode token, expected base64: %s", err)
	}

	opts := token{}
	encoder := gob.NewDecoder(bytes.NewReader(decoding))
	if err := encoder.Decode(&opts); err != nil {
		return token{}, fmt.Errorf("failed to decode token bytes: %s", err)
	}

	return opts, nil
}
