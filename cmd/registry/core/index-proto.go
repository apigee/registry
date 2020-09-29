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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/gogo/protobuf/jsonpb"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

func NewIndexFromZippedProtos(b []byte) (*rpc.Index, error) {
	// create a tmp directory
	dname, err := ioutil.TempDir("", "registry-protos-")
	if err != nil {
		return nil, err
	}
	// whenever we finish, delete the tmp directory
	defer os.RemoveAll(dname)
	// unzip the protos to the temp directory
	_, err = Unzip(b, dname)
	if err != nil {
		return nil, err
	}
	// process the directory
	c, err := indexForRoot(dname)
	if err != nil {
		return nil, err
	}
	BuildIndex(c)
	RemoveRequestAndResponseSchemas(c)
	FlattenPaths(c)
	return c, nil
}

func indexForRoot(root string) (*rpc.Index, error) {
	s := &rpc.Index{}
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				file, err := fileForProto(path, root+"/")
				if err != nil {
					return err
				}
				s.Files = append(s.Files, file)
			}
			return nil
		})
	return s, err
}

func fileForProto(filename, root string) (*rpc.File, error) {
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
		return nil, err
	}
	v, err := protoparser.UnorderedInterpret(p)
	if err != nil {
		return nil, err
	}
	f := &rpc.File{}
	f.FileName = strings.TrimPrefix(filename, root)
	for _, m := range v.ProtoBody.Messages {
		f.Schemas = append(f.Schemas, schemaForMessage(m))
	}
	for _, s := range v.ProtoBody.Services {
		serviceName := s.ServiceName
		for _, rpc := range s.ServiceBody.RPCs {
			f.Operations = append(f.Operations, operationForRPC(rpc, serviceName))
		}
	}
	return f, nil
}

func schemaForMessage(m *unordered.Message) *rpc.Schema {
	s := &rpc.Schema{}
	s.SchemaName = m.MessageName
	for _, f := range m.MessageBody.Fields {
		field := &rpc.Field{}
		field.FieldName = f.FieldName
		s.Fields = append(s.Fields, field)
	}
	for _, opt := range m.MessageBody.Options {
		processOptionForSchema(s, opt)
	}
	return s
}

func linesForOption(opt *parser.Option) []string {
	body := opt.Constant
	body = strings.TrimPrefix(body, "{")
	body = strings.TrimSuffix(body, "}")
	body = strings.ReplaceAll(body, ",", "\n")
	return strings.Split(body, "\n")
}

func processOptionForSchema(s *rpc.Schema, opt *parser.Option) {
	if opt.OptionName == "(google.api.resource)" {
		lines := linesForOption(opt)
		for _, line := range lines {
			pair := strings.SplitN(line, ":", 2)
			if len(pair) == 2 {
				name := pair[0]
				if name == "type" {
					s.ResourceType = trimQuotes(pair[1])
				}
				if name == "pattern" {
					s.ResourceName = trimQuotes(pair[1])
				}
			}
		}
	}
}

func operationForRPC(myrpc *parser.RPC, serviceName string) *rpc.Operation {
	op := &rpc.Operation{}
	op.OperationName = myrpc.RPCName
	op.ServiceName = serviceName
	for _, opt := range myrpc.Options {
		processOptionForOperation(op, opt)
	}
	return op
}

func processOptionForOperation(op *rpc.Operation, opt *parser.Option) {
	if opt.OptionName == "(google.api.http)" {
		lines := linesForOption(opt)
		for _, line := range lines {
			pair := strings.SplitN(line, ":", 2)
			if len(pair) == 2 {
				name := pair[0]
				if name == "get" || name == "post" || name == "put" || name == "delete" || name == "patch" {
					op.Verb = name
					op.Path = trimQuotes(pair[1])
				}
			}
		}
	}
}

func trimQuotes(s string) string {
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")
	return s
}

// FlattenPaths removes assignments and parameters from operation paths
func FlattenPaths(index *rpc.Index) {
	r1 := regexp.MustCompile("{[^{}=]+=([^{}=]+)}")
	r2 := regexp.MustCompile("{[^{}].*}")
	for _, op := range index.Operations {
		p := op.Path
		p = strings.ReplaceAll(p, "{$api_version}", "v*")
		p = r1.ReplaceAllString(p, "$1")
		p = r2.ReplaceAllString(p, "*")
		op.Path = p
	}
}

// BuildIndex adds flat arrays of fields, schemas, and operations.
func BuildIndex(index *rpc.Index) {
	index.Fields = make([]*rpc.Field, 0)
	index.Schemas = make([]*rpc.Schema, 0)
	index.Operations = make([]*rpc.Operation, 0)
	for _, file := range index.Files {
		for _, op := range file.Operations {
			opCopy := &rpc.Operation{}
			*opCopy = *op
			opCopy.FileName = file.FileName
			index.Operations = append(index.Operations, opCopy)
		}
		for _, schema := range file.Schemas {
			schemaCopy := &rpc.Schema{}
			*schemaCopy = *schema
			schemaCopy.Fields = nil
			schemaCopy.FileName = file.FileName
			index.Schemas = append(index.Schemas, schemaCopy)
			for _, field := range schema.Fields {
				fieldCopy := &rpc.Field{}
				*fieldCopy = *field
				fieldCopy.FileName = file.FileName
				fieldCopy.SchemaName = schema.SchemaName
				index.Fields = append(index.Fields, fieldCopy)
			}
		}
	}
	sort.Slice(index.Fields, func(i, j int) bool {
		return index.Fields[i].FieldName < index.Fields[j].FieldName
	})
	sort.Slice(index.Schemas, func(i, j int) bool {
		return index.Schemas[i].SchemaName < index.Schemas[j].SchemaName
	})
	sort.Slice(index.Operations, func(i, j int) bool {
		return index.Operations[i].OperationName < index.Operations[j].OperationName
	})
}

// RemoveRequestAndResponseSchemas removes these from the flat schema list.
func RemoveRequestAndResponseSchemas(index *rpc.Index) {
	filteredSchemas := make([]*rpc.Schema, 0)
	for _, schema := range index.Schemas {
		if strings.HasSuffix(schema.SchemaName, "Request") ||
			strings.HasSuffix(schema.SchemaName, "Response") {
			// skip it
		} else {
			filteredSchemas = append(filteredSchemas, schema)
		}
	}
	index.Schemas = filteredSchemas
}

// ExportSchemas writes an index of Schemas as a CSV
func ExportSchemas(index *rpc.Index) error {
	f, err := os.Create("schemas.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, schema := range index.Schemas {
		fmt.Fprintf(w, "%s,%s,%s,%s\n",
			schema.SchemaName, schema.ResourceName, schema.ResourceType, schema.FileName)
	}
	w.Flush()
	return nil
}

// ExportOperations writes an index of Operations as a CSV
func ExportOperations(index *rpc.Index) error {
	f, err := os.Create("operations.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, op := range index.Operations {
		fmt.Fprintf(w, "%s,%s,%s,%s,%s\n",
			op.OperationName, op.ServiceName, op.Verb, op.Path, op.FileName)
	}
	w.Flush()
	return nil
}

// ExportFields writes an index of Fields as a CSV
func ExportFields(index *rpc.Index) error {
	f, err := os.Create("fields.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, field := range index.Fields {
		fmt.Fprintf(w, "%s,%s,%s\n",
			field.FieldName, field.SchemaName, field.FileName)
	}
	w.Flush()
	return nil
}

// ExportAsJSON writes a index as a JSON file
func ExportAsJSON(index *rpc.Index) error {
	f, err := os.Create("index.json")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	m := &jsonpb.Marshaler{
		Indent: "  ",
	}
	err = m.Marshal(w, index)
	w.Flush()
	return err
}
