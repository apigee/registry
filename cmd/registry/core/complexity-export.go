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
	"fmt"
	"log"
	"strings"

	"github.com/apigee/registry/rpc"
	metrics "github.com/googleapis/gnostic/metrics"
	"google.golang.org/protobuf/proto"
)

func ExportComplexityToSheet(name string, inputs []*rpc.Property) error {
	sheetsClient, err := NewSheetsClient("")
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	sheet, err := sheetsClient.CreateSheet(name, []string{"Complexity"})
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	_, err = sheetsClient.FormatHeaderRow(sheet.Sheets[0].Properties.SheetId)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	rows := make([][]interface{}, 0)
	rows = append(rows, rowForLabeledComplexity("", "", nil))
	for _, input := range inputs {
		complexity, err := getComplexity(input)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		parts := strings.Split(input.Name, "/") // use to get api_id [3] and version_id [5]
		rows = append(rows, rowForLabeledComplexity(parts[3], parts[5], complexity))
	}
	_, err = sheetsClient.Update(fmt.Sprintf("Complexity"), rows)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	log.Printf("exported to %+v\n", sheet.SpreadsheetUrl)
	return nil
}

func getComplexity(property *rpc.Property) (*metrics.Complexity, error) {
	switch v := property.GetValue().(type) {
	case *rpc.Property_MessageValue:
		if v.MessageValue.TypeUrl == "gnostic.metrics.Complexity" {
			value := &metrics.Complexity{}
			err := proto.Unmarshal(v.MessageValue.Value, value)
			if err != nil {
				return nil, err
			} else {
				return value, nil
			}
		} else {
			return nil, fmt.Errorf("not a complexity: %s", property.Name)
		}
	default:
		return nil, fmt.Errorf("not a complexity: %s", property.Name)
	}
}

func rowForLabeledComplexity(api, version string, c *metrics.Complexity) []interface{} {
	if c == nil {
		return []interface{}{
			"api",
			"version",
			"schemas",
			"schema properties",
			"paths",
			"gets",
			"posts",
			"puts",
			"deletes",
		}
	}
	return []interface{}{api, version, c.SchemaCount, c.SchemaPropertyCount, c.PathCount, c.GetCount, c.PostCount, c.PutCount, c.DeleteCount}
}
