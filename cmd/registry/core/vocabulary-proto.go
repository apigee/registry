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
	"sort"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"

	metrics "github.com/googleapis/gnostic/metrics"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
)

func NewVocabularyFromZippedProtos(b []byte) (*metrics.Vocabulary, error) {
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
	return vocabularyForPath(dname)
}

func vocabularyForPath(path string) (*metrics.Vocabulary, error) {
	v := NewVocabulary()
	err := v.fillVocabularyFromPath(path)
	if err != nil {
		return nil, err
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

// NewVocabulary creates a new Vocabulary object.
func NewVocabulary() *Vocabulary {
	return &Vocabulary{
		Schemas:    make(map[string]int),
		Operations: make(map[string]int),
		Parameters: make(map[string]int),
		Properties: make(map[string]int),
	}
}

func (vocab *Vocabulary) fillVocabularyFromPath(path string) error {
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				err := vocab.fillVocabularyFromProto(path)
				if err != nil {
					return err
				}
			}
			return nil
		})
	return err
}

func (vocab *Vocabulary) fillVocabularyFromProto(filename string) error {
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
			vocab.fillVocabularyFromMessage(m)
		case *parser.Service:
			vocab.fillVocabularyFromService(m)
		default:
			// fmt.Printf("IGNORED %+v\n", v)
		}
	}
	return nil
}

func (vocab *Vocabulary) fillVocabularyFromMessage(m *parser.Message) {
	vocab.Schemas[m.MessageName]++

	for _, x := range m.MessageBody {
		switch v := x.(type) {
		case *parser.Field:
			vocab.Properties[v.FieldName]++
		default:
			// fmt.Printf("IGNORED %+v\n", v)
		}
	}
}

func (vocab *Vocabulary) fillVocabularyFromService(m *parser.Service) {
	for _, x := range m.ServiceBody {
		switch v := x.(type) {
		case *parser.RPC:
			vocab.Operations[v.RPCName]++
		default:
			// fmt.Printf("IGNORED %+v\n", v)
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
