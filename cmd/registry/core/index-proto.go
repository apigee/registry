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
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	protoparser "github.com/yoheimuta/go-protoparser/v4"
)

func NewIndexFromZippedProtos(b []byte) (*rpc.Index, error) {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	s := &rpc.Index{}
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

		v, err := protoparser.UnorderedInterpret(p)
		if err != nil {
			return nil, err
		}

		file := &rpc.File{
			Name: f.Name,
		}

		for _, m := range v.ProtoBody.Messages {
			file.Schemas = append(file.Schemas, schemaForMessage(m))
		}

		for _, s := range v.ProtoBody.Services {
			for _, rpc := range s.ServiceBody.RPCs {
				file.Operations = append(file.Operations, operationForRPC(rpc, s.ServiceName))
			}
		}

		s.Files = append(s.Files, file)
	}

	buildIndex(s)
	removeRequestAndResponseSchemas(s)
	flattenPaths(s)
	return s, nil
}

func schemaForMessage(m *unordered.Message) *rpc.Schema {
	s := &rpc.Schema{}
	s.Name = m.MessageName
	for _, f := range m.MessageBody.Fields {
		field := &rpc.Field{}
		field.Name = f.FieldName
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
					s.Type = trimQuotes(pair[1])
				}
				if name == "pattern" {
					s.Name = trimQuotes(pair[1])
				}
			}
		}
	}
}

func operationForRPC(myrpc *parser.RPC, serviceName string) *rpc.Operation {
	op := &rpc.Operation{}
	op.Name = myrpc.RPCName
	op.Service = serviceName
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

// flattenPaths removes assignments and parameters from operation paths
func flattenPaths(index *rpc.Index) {
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

// buildIndex adds flat arrays of fields, schemas, and operations.
func buildIndex(index *rpc.Index) {
	index.Fields = make([]*rpc.Field, 0)
	index.Schemas = make([]*rpc.Schema, 0)
	index.Operations = make([]*rpc.Operation, 0)
	for _, file := range index.Files {
		for _, op := range file.Operations {
			index.Operations = append(index.Operations, &rpc.Operation{
				Name:    op.GetName(),
				Service: op.GetService(),
				Verb:    op.GetVerb(),
				Path:    op.GetPath(),
				File:    file.GetName(),
			})
		}
		for _, schema := range file.Schemas {
			index.Schemas = append(index.Schemas, &rpc.Schema{
				Name:     schema.GetName(),
				Resource: schema.GetResource(),
				Type:     schema.GetType(),
				File:     file.GetName(),
				Fields:   nil,
			})
			for _, field := range schema.Fields {
				index.Fields = append(index.Fields, &rpc.Field{
					Name:   field.GetName(),
					Schema: schema.GetName(),
					File:   file.GetName(),
				})
			}
		}
	}
	sort.Slice(index.Fields, func(i, j int) bool {
		return index.Fields[i].Name < index.Fields[j].Name
	})
	sort.Slice(index.Schemas, func(i, j int) bool {
		return index.Schemas[i].Name < index.Schemas[j].Name
	})
	sort.Slice(index.Operations, func(i, j int) bool {
		return index.Operations[i].Name < index.Operations[j].Name
	})
}

// removeRequestAndResponseSchemas removes these from the flat schema list.
func removeRequestAndResponseSchemas(index *rpc.Index) {
	filteredSchemas := make([]*rpc.Schema, 0)
	for _, schema := range index.Schemas {
		if strings.HasSuffix(schema.Name, "Request") ||
			strings.HasSuffix(schema.Name, "Response") {
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
			schema.Name, schema.Resource, schema.Type, schema.File)
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
			op.Name, op.Service, op.Verb, op.Path, op.File)
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
			field.Name, field.Schema, field.File)
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
