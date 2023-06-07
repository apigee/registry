// Copyright 2020 Google LLC.
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

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ValidateMask returns an error if and only if the given mask does not follow AIP-134 guidance.
func ValidateMask(message protoreflect.ProtoMessage, mask *fieldmaskpb.FieldMask) error {
	// Nil masks are valid for implicit updates.
	if mask == nil {
		return nil
	}

	// Wildcard masks are valid for full replacements.
	if len(mask.GetPaths()) == 1 && mask.Paths[0] == "*" {
		return nil
	}

	// Other masks are valid if all their fields exist on the message.
	unknowns := make([]string, 0)
	for _, field := range mask.GetPaths() {
		// New returns an error if the field isn't valid for the message.
		if _, err := fieldmaskpb.New(message, field); err != nil {
			unknowns = append(unknowns, field)
		}
	}

	if len(unknowns) > 0 {
		return fmt.Errorf("unrecognized fields %v", unknowns)
	}

	return nil
}

// ExpandMask returns a field mask for the given message following AIP-134 guidance.
// When the mask argument is nil, only populated (non-default) fields are included.
// When the mask argument only contains the wildcard character '*', every proto fields is included.
func ExpandMask(m protoreflect.ProtoMessage, mask *fieldmaskpb.FieldMask) *fieldmaskpb.FieldMask {
	if mask == nil || len(mask.GetPaths()) == 0 {
		return populatedFields(m)
	} else if len(mask.GetPaths()) == 1 && mask.Paths[0] == "*" {
		return allFields(m)
	}

	mask.Normalize()
	return mask
}

func populatedFields(m protoreflect.ProtoMessage) *fieldmaskpb.FieldMask {
	// New does not return errors unless path arguments are provided.
	mask, _ := fieldmaskpb.New(m)

	// Range iterates over the populated proto fields only.
	m.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		_ = mask.Append(m, string(field.Name()))
		return true // Continue iterating.
	})

	return mask
}

func allFields(m protoreflect.ProtoMessage) *fieldmaskpb.FieldMask {
	// New does not return errors unless path arguments are provided.
	mask, _ := fieldmaskpb.New(m)

	// Fields gives us all of the proto's field names.
	fields := m.ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		_ = mask.Append(m, string(fields.Get(i).Name()))
	}

	return mask
}
