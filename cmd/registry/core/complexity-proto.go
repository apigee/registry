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

package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"

	metrics "github.com/googleapis/gnostic/metrics"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
)

func NewComplexityFromZippedProtos(b []byte) (*metrics.Complexity, error) {
	// create a tmp directory
	dname, err := ioutil.TempDir("", "registry-protos-")
	if err != nil {
		return nil, err
	}
	// whenever we finish, delete the tmp directory
	defer os.RemoveAll(dname)
	// unzip the protos to the temp directory
	_, err = unzipArchiveToPath(b, dname)
	if err != nil {
		return nil, err
	}
	// process the directory
	return complexityForPath(dname)
}

func complexityForPath(path string) (*metrics.Complexity, error) {
	c := &metrics.Complexity{}
	err := fillComplexityFromPath(c, path)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func fillComplexityFromPath(c *metrics.Complexity, path string) error {
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				err := fillComplexityFromProto(c, path)
				if err != nil {
					return err
				}
			}
			return nil
		})
	return err
}

func fillComplexityFromProto(c *metrics.Complexity, filename string) error {
	reader, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer reader.Close()

	p, err := protoparser.Parse(
		reader,
		protoparser.WithDebug(false),
		protoparser.WithPermissive(true),
		protoparser.WithFilename(filepath.Base(filename)),
	)
	if err != nil {
		return err
	}

	for _, x := range p.ProtoBody {
		switch m := x.(type) {
		case *parser.Message:
			fillComplexityFromMessage(c, m)
		case *parser.Service:
			fillComplexityFromService(c, m)
		default:
			// fmt.Printf("IGNORED %+v\n", v)
		}
	}
	return nil
}

func fillComplexityFromMessage(c *metrics.Complexity, m *parser.Message) {
	c.SchemaCount++
	for _, x := range m.MessageBody {
		switch x.(type) {
		case *parser.Field:
			c.SchemaPropertyCount++
		default:
			// fmt.Printf("IGNORED %+v\n", v)
		}
	}
}

func fillComplexityFromService(c *metrics.Complexity, m *parser.Service) {
	for _, x := range m.ServiceBody {
		switch x.(type) {
		case *parser.RPC:
			c.PathCount++
		default:
			// fmt.Printf("IGNORED %+v\n", v)
		}
	}
}
