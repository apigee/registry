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

package vocabulary

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"

	metrics "github.com/google/gnostic/metrics"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
)

func NewVocabularyFromZippedProtos(b []byte) (*metrics.Vocabulary, error) {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	v := &Vocabulary{
		Schemas:    make(map[string]int),
		Operations: make(map[string]int),
		Parameters: make(map[string]int),
		Properties: make(map[string]int),
	}

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
			switch m := x.(type) {
			case *parser.Message:
				v.fillVocabularyFromMessage(m)
			case *parser.Service:
				v.fillVocabularyFromService(m)
			}
		}
	}

	return &metrics.Vocabulary{
		Properties: fillProtoStructure(v.Properties),
		Schemas:    fillProtoStructure(v.Schemas),
		Operations: fillProtoStructure(v.Operations),
		Parameters: fillProtoStructure(v.Parameters),
	}, nil
}

// Vocabulary represents the counts of various types of terms in an API.
type Vocabulary struct {
	Schemas    map[string]int
	Operations map[string]int
	Parameters map[string]int
	Properties map[string]int
}

func (vocab *Vocabulary) fillVocabularyFromMessage(m *parser.Message) {
	vocab.Schemas[m.MessageName]++

	for _, x := range m.MessageBody {
		if v, ok := x.(*parser.Field); ok {
			vocab.Properties[v.FieldName]++
		}
	}
}

func (vocab *Vocabulary) fillVocabularyFromService(m *parser.Service) {
	for _, x := range m.ServiceBody {
		if v, ok := x.(*parser.RPC); ok {
			vocab.Operations[v.RPCName]++
		}
	}
}

// fillProtoStructure adds data to the Word Count structure.
// The Word Count structure can then be added to the Vocabulary protocol buffer.
func fillProtoStructure(m map[string]int) []*metrics.WordCount {
	keyNames := make([]string, 0, len(m))
	for key := range m {
		keyNames = append(keyNames, key)
	}
	sort.Strings(keyNames)

	counts := make([]*metrics.WordCount, 0)
	for _, k := range keyNames {
		temp := &metrics.WordCount{
			Word:  k,
			Count: int32(m[k]),
		}
		counts = append(counts, temp)
	}
	return counts
}
