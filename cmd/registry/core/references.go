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
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// NewReferencesFromOpenAPIv2 ...
func NewReferencesFromOpenAPIv2(document *openapi_v2.Document) (*rpc.References, error) {
	return nil, nil
}

// NewReferencesFromOpenAPIv3 ...
func NewReferencesFromOpenAPIv3(document *openapi_v3.Document) (*rpc.References, error) {
	return nil, nil
}

// NewReferencesFromZippedProtos ...
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
	internals, externals, err := internalsAndExternalsForPath(dname)
	externals = removeInternalsAndDuplicatesFromExternals(internals, externals)
	sort.Strings(internals)
	sort.Strings(externals)
	return &rpc.References{AvailableReferences: internals, ExternalReferences: externals}, err
}

func internalsAndExternalsForPath(root string) ([]string, []string, error) {
	internals := []string{}
	externals := []string{}
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				name := strings.TrimPrefix(path, root+"/")
				internals = append(internals, name)
				externals, err = fillExternalsFromProto(externals, path)
				if err != nil {
					return err
				}
			}
			return nil
		})
	return internals, externals, err
}

func fillExternalsFromProto(externals []string, filename string) ([]string, error) {
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
		return externals, err
	}

	for _, x := range p.ProtoBody {
		switch m := x.(type) {
		case *parser.Import:
			externals = append(externals, strings.Trim(m.Location, "\""))
		default:
		}
	}
	return externals, nil
}

func removeInternalsAndDuplicatesFromExternals(internals, externals []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, e := range internals {
		keys[e] = true
	}
	for _, d := range externals {
		if !keys[d] {
			keys[d] = true
			list = append(list, d)
		}
	}
	return list
}
