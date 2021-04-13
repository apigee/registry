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

package dao

import (
	"github.com/apigee/registry/server/storage"
)

// PageOptions contains custom arguments for listing requests.
type PageOptions struct {
	// Size is the maximum number of resources to include in the response.
	// If unspecified, page size defaults to 50. Values above 1000 are coerced to 1000.
	Size int32
	// Filter is the filter string for this listing request, as described at https://google.aip.dev/160.
	Filter string
	// Token is a value returned from with a previous page in a series of listing requests.
	// If specified, listing will continue from the end of the previous page. Otherwise,
	// the first page in a listing series will be returned.
	Token string
}

type DAO struct {
	storage.Client
}

func NewDAO(c storage.Client) DAO {
	return DAO{
		Client: c,
	}
}
