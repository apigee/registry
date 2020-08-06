package corpus

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	protoparser "github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// ReadCorpus reads an API corpus from a file or directory.
func ReadCorpus(path string) (*Corpus, error) {
	if fi, err := os.Stat(path); err == nil {
		if fi.Mode().IsRegular() {
			return corpusForFile(path)
		}
		if fi.Mode().IsDir() {
			return corpusForPath(path)
		}
	}
	return nil, fmt.Errorf("%s not found", path)
}

// BuildIndex adds flat arrays of fields, schemas, and operations.
func (corpus *Corpus) BuildIndex() {
	corpus.Fields = make([]*Field, 0)
	corpus.Schemas = make([]*Schema, 0)
	corpus.Operations = make([]*Operation, 0)
	for _, file := range corpus.Files {
		for _, op := range file.Operations {
			opCopy := &Operation{}
			*opCopy = *op
			opCopy.FileName = file.FileName
			corpus.Operations = append(corpus.Operations, opCopy)
		}
		for _, schema := range file.Schemas {
			schemaCopy := &Schema{}
			*schemaCopy = *schema
			schemaCopy.Fields = nil
			schemaCopy.FileName = file.FileName
			corpus.Schemas = append(corpus.Schemas, schemaCopy)
			for _, field := range schema.Fields {
				fieldCopy := &Field{}
				*fieldCopy = *field
				fieldCopy.FileName = file.FileName
				fieldCopy.SchemaName = schema.SchemaName
				corpus.Fields = append(corpus.Fields, fieldCopy)
			}
		}
	}
	sort.Slice(corpus.Fields, func(i, j int) bool {
		return corpus.Fields[i].FieldName < corpus.Fields[j].FieldName
	})
	sort.Slice(corpus.Schemas, func(i, j int) bool {
		return corpus.Schemas[i].SchemaName < corpus.Schemas[j].SchemaName
	})
	sort.Slice(corpus.Operations, func(i, j int) bool {
		return corpus.Operations[i].OperationName < corpus.Operations[j].OperationName
	})
}

// ExportSchemas writes an index of Schemas as a CSV
func (corpus *Corpus) ExportSchemas() error {
	f, err := os.Create("schemas.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, schema := range corpus.Schemas {
		fmt.Fprintf(w, "%s,%s,%s,%s\n",
			schema.SchemaName, schema.ResourceName, schema.ResourceType, schema.FileName)
	}
	w.Flush()
	return nil
}

// ExportOperations writes an index of Operations as a CSV
func (corpus *Corpus) ExportOperations() error {
	f, err := os.Create("operations.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, op := range corpus.Operations {
		fmt.Fprintf(w, "%s,%s,%s,%s,%s\n",
			op.OperationName, op.ServiceName, op.Verb, op.Path, op.FileName)
	}
	w.Flush()
	return nil
}

// ExportFields writes an index of Fields as a CSV
func (corpus *Corpus) ExportFields() error {
	f, err := os.Create("fields.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, field := range corpus.Fields {
		fmt.Fprintf(w, "%s,%s,%s\n",
			field.FieldName, field.SchemaName, field.FileName)
	}
	w.Flush()
	return nil
}

// ExportAsJSON writes a corpus as a JSON file
func (corpus *Corpus) ExportAsJSON() error {
	f, err := os.Create("corpus.json")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	m := &jsonpb.Marshaler{
		Indent: "  ",
	}
	err = m.Marshal(w, corpus)
	w.Flush()
	return err
}

func corpusForFile(path string) (*Corpus, error) {
	f, err := fileForProto(path)
	if err != nil {
		return nil, err
	}
	return &Corpus{Files: []*File{f}}, nil
}

func corpusForPath(path string) (*Corpus, error) {
	s := &Corpus{}
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				file, err := fileForProto(path)
				if err != nil {
					return err
				}
				s.Files = append(s.Files, file)
			}
			return nil
		})
	return s, err
}

func fileForProto(filename string) (*File, error) {
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
	f := &File{}
	f.FileName = filename
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

func schemaForMessage(m *unordered.Message) *Schema {
	s := &Schema{}
	s.SchemaName = m.MessageName
	for _, f := range m.MessageBody.Fields {
		field := &Field{}
		field.FieldName = f.FieldName
		s.Fields = append(s.Fields, field)
	}
	for _, opt := range m.MessageBody.Options {
		s.processOption(opt)
	}
	return s
}

func linesForOption(opt *parser.Option) []string {
	body := opt.Constant
	body = strings.TrimPrefix(body, "{")
	body = strings.TrimSuffix(body, "}")
	return strings.Split(body, "\n")
}

func (s *Schema) processOption(opt *parser.Option) {
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

func operationForRPC(rpc *parser.RPC, serviceName string) *Operation {
	op := &Operation{}
	op.OperationName = rpc.RPCName
	op.ServiceName = serviceName
	for _, opt := range rpc.Options {
		op.processOption(opt)
	}
	return op
}

func (op *Operation) processOption(opt *parser.Option) {
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
