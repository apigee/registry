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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apigee/registry/rpc"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// NewReferencesFromZippedProtos computes references of a Protobuf spec.
func NewReferencesFromZippedProtos(b []byte) (*rpc.References, error) {
	// create a tmp directory
	dname, err := ioutil.TempDir("", "registry-protos-")
	if err != nil {
		return nil, err
	}
	// whenever we finish, delete the tmp directory
	defer os.RemoveAll(dname)
	// unzip the protos to the temp directory
	_, err = UnzipArchiveToPath(b, dname)
	if err != nil {
		return nil, err
	}
	// process the directory
	references, files, err := collectReferencesAndFilesForPath(dname)
	references = filterFilesAndDuplicatesFromReferences(files, references)
	sort.Strings(files)
	sort.Strings(references)
	return &rpc.References{AvailableReferences: files, ExternalReferences: references}, err
}

// collectReferencesAndFilesForPath builds lists of external references and internal
// files for a set of files in a directory corresponding to a protobuf API spec.
func collectReferencesAndFilesForPath(root string) ([]string, []string, error) {
	references := []string{}
	files := []string{}
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				name := strings.TrimPrefix(path, root+"/")
				files = append(files, name)
				references, err = collectReferencesForProto(references, path)
				if err != nil {
					return err
				}
			}
			return nil
		})
	return references, files, err
}

// collectReferencesForProto fills a slice with references found in the import statements in a proto file.
func collectReferencesForProto(references []string, filename string) ([]string, error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	p, err := protoparser.Parse(
		reader,
		protoparser.WithDebug(false),
		protoparser.WithPermissive(true),
		protoparser.WithFilename(filepath.Base(filename)),
	)
	if err != nil {
		return references, err
	}

	for _, x := range p.ProtoBody {
		if m, ok := x.(*parser.Import); ok {
			references = append(references, strings.Trim(m.Location, "\""))
		}
	}
	return references, nil
}

// filterFilesAndDuplicatesFromReferences removes internal references
// (other files in the same spec) and duplicates from the list of externals.
func filterFilesAndDuplicatesFromReferences(files, references []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, e := range files {
		keys[e] = true
	}
	for _, d := range references {
		if !keys[d] {
			keys[d] = true
			list = append(list, d)
		}
	}
	return list
}
