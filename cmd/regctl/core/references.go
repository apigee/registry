// Copyright 2021 Google LLC. All Rights Reserved.
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

package core

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	protoparser "github.com/yoheimuta/go-protoparser/v4"
)

// NewReferencesFromZippedProtos computes references of a Protobuf spec.
func NewReferencesFromZippedProtos(b []byte) (*rpc.References, error) {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)
	fileSet := make(map[string]bool)
	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, ".proto") {
			continue
		}

		fileSet[f.Name] = true
		files = append(files, f.Name)
	}

	refs := make([]string, 0)
	refSet := make(map[string]bool)
	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, ".proto") {
			continue
		}

		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer r.Close()

		opts := []protoparser.Option{
			protoparser.WithDebug(false),
			protoparser.WithPermissive(true),
			protoparser.WithFilename(filepath.Base(f.Name)),
		}

		p, err := protoparser.Parse(r, opts...)
		if err != nil {
			return nil, err
		}

		for _, x := range p.ProtoBody {
			if ref, ok := x.(*parser.Import); ok {
				ref := strings.Trim(ref.Location, "\"") // Remove quotation marks from references.
				if _, ok := refSet[ref]; ok {
					continue // Ignore references we've already listed.
				}

				if _, ok := fileSet[ref]; ok {
					continue // Ignore references to other files in the spec.
				}

				refSet[ref] = true
				refs = append(refs, ref)
			}
		}
	}

	sort.Strings(files)
	sort.Strings(refs)
	return &rpc.References{
		AvailableReferences: files,
		ExternalReferences:  refs,
	}, err
}
