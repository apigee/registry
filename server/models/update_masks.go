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

import "google.golang.org/protobuf/types/known/fieldmaskpb"

// activeUpdateMask returns true if an update mask should be used to filter fields of a message.
func activeUpdateMask(mask *fieldmaskpb.FieldMask) bool {
	if mask == nil {
		return false
	}
	if len(mask.Paths) == 0 {
		return false
	}
	if len(mask.Paths) == 1 && mask.Paths[0] == "*" {
		return false
	}
	return true
}
