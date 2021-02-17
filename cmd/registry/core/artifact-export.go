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
	"strconv"

	"github.com/apigee/registry/rpc"
)

func ExportInt64ToSheet(name string, artifacts []*rpc.Artifact) (string, error) {
	sheetsClient, err := NewSheetsClient("")
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	sheet, err := sheetsClient.CreateSheet(name, []string{"Values"})
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	_, err = sheetsClient.FormatHeaderRow(sheet.Sheets[0].Properties.SheetId)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	rows := make([][]interface{}, 0)
	rows = append(rows, rowForInt64Artifact(nil))
	for _, artifact := range artifacts {
		rows = append(rows, rowForInt64Artifact(artifact))
	}
	_, err = sheetsClient.Update("Values", rows)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	return sheet.SpreadsheetUrl, nil
}

func rowForInt64Artifact(artifact *rpc.Artifact) []interface{} {
	if artifact == nil {
		return []interface{}{
			"name",
			"value",
		}
	}

	contents := artifact.GetContents()
	value, _ := strconv.ParseInt(string(contents), 10, 64)
	subject := artifact.GetName()
	return []interface{}{subject, value}
}
