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

package complexity

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"

	metrics "github.com/google/gnostic/metrics"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
)

func SummarizeZippedProtos(b []byte) (*metrics.Complexity, error) {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	c := &metrics.Complexity{}
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

		c.SchemaCount += messageCount(p)
		for _, x := range p.ProtoBody {
			switch t := x.(type) {
			case *parser.Message:
				c.SchemaPropertyCount += fieldCount(t)
			case *parser.Service:
				c.PathCount += rpcCount(t)
			}
		}
	}

	return c, nil
}

func messageCount(p *parser.Proto) (count int32) {
	for _, x := range p.ProtoBody {
		if _, ok := x.(*parser.Message); ok {
			count++
		}
	}
	return
}

func fieldCount(m *parser.Message) (count int32) {
	for _, x := range m.MessageBody {
		if _, ok := x.(*parser.Field); ok {
			count++
		}
	}
	return
}

func rpcCount(s *parser.Service) (count int32) {
	for _, x := range s.ServiceBody {
		if _, ok := x.(*parser.RPC); ok {
			count++
		}
	}
	return
}
