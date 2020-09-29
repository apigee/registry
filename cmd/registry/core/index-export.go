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
	"log"

	"github.com/apigee/registry/rpc"
)

func ExportIndexToSheet(name string, index *rpc.Index) error {
	sheetsClient, err := NewSheetsClient("")
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	sheet, err := sheetsClient.CreateSheet(name, []string{"Operations", "Schemas", "Fields"})
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	for _, s := range sheet.Sheets {
		_, err := sheetsClient.FormatHeaderRow(s.Properties.SheetId)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	}
	{
		rows := make([][]interface{}, 0)
		rows = append(rows, rowForOperation(nil))
		for _, op := range index.Operations {
			rows = append(rows, rowForOperation(op))
		}
		_, err = sheetsClient.Update("Operations", rows)
	}
	{
		rows := make([][]interface{}, 0)
		rows = append(rows, rowForSchema(nil))
		for _, op := range index.Schemas {
			rows = append(rows, rowForSchema(op))
		}
		_, err = sheetsClient.Update("Schemas", rows)
	}
	{
		rows := make([][]interface{}, 0)
		rows = append(rows, rowForField(nil))
		for _, op := range index.Fields {
			rows = append(rows, rowForField(op))
		}
		_, err = sheetsClient.Update("Fields", rows)
	}
	log.Printf("exported to %+v\n", sheet.SpreadsheetUrl)
	return nil
}

func rowForOperation(op *rpc.Operation) []interface{} {
	if op == nil {
		return []interface{}{"rpc", "service", "verb", "path", "file"}
	}
	row := make([]interface{}, 0)
	row = append(row, op.OperationName)
	row = append(row, op.ServiceName)
	row = append(row, op.Verb)
	row = append(row, op.Path)
	row = append(row, op.FileName)
	return row
}

func rowForSchema(s *rpc.Schema) []interface{} {
	if s == nil {
		return []interface{}{"message", "resource name", "type", "file"}
	}
	row := make([]interface{}, 0)
	row = append(row, s.SchemaName)
	row = append(row, s.ResourceName)
	row = append(row, s.ResourceType)
	row = append(row, s.FileName)
	return row
}

func rowForField(f *rpc.Field) []interface{} {
	if f == nil {
		return []interface{}{"field", "message", "file"}
	}
	row := make([]interface{}, 0)
	row = append(row, f.FieldName)
	row = append(row, f.SchemaName)
	row = append(row, f.FileName)
	return row
}
