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

func ExportVersionHistoryToSheet(name string, property *rpc.Property) error {
	sheetsClient, err := NewSheetsClient("")
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	versionHistory, err := getVersionHistory(property)
	sheetNames := []string{"Summary"}
	for _, version := range versionHistory.Versions {
		versionName := nameForVersion(version.Name)
		sheetNames = append(sheetNames, versionName+"-new")
		sheetNames = append(sheetNames, versionName+"-deleted")
	}
	sheet, err := sheetsClient.CreateSheet(name, sheetNames)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	rows := make([][]interface{}, 0)
	rows = append(rows, rowForVersionSummary(nil))
	for _, version := range versionHistory.Versions {
		rows = append(rows, rowForVersionSummary(version))
	}
	_, err = sheetsClient.Update(fmt.Sprintf("Summary"), rows)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	for _, version := range versionHistory.Versions {
		versionName := nameForVersion(version.Name)
		rows := rowsForVocabulary(version.NewTerms)
		_, err = sheetsClient.Update(fmt.Sprintf(versionName+"-new"), rows)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		rows = rowsForVocabulary(version.DeletedTerms)
		_, err = sheetsClient.Update(fmt.Sprintf(versionName+"-deleted"), rows)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	}
	log.Printf("exported to %+v\n", sheet.SpreadsheetUrl)
	return nil
}

func nameForVersion(version string) string {
	parts := strings.Split(version, "/")
	return parts[5]
}

func getVersionHistory(property *rpc.Property) (*metrics.VersionHistory, error) {
	switch v := property.GetValue().(type) {
	case *rpc.Property_MessageValue:
		if v.MessageValue.TypeUrl == "gnostic.metrics.VersionHistory" {
			value := &metrics.VersionHistory{}
			err := proto.Unmarshal(v.MessageValue.Value, value)
			if err != nil {
				return nil, err
			} else {
				return value, nil
			}
		} else {
			return nil, fmt.Errorf("not a version history: %s", property.Name)
		}
	default:
		return nil, fmt.Errorf("not a version history: %s", property.Name)
	}
}

func rowForVersionSummary(v *metrics.Version) []interface{} {
	if v == nil {
		return []interface{}{
			"version",
			"new terms",
			"deleted terms",
		}
	}
	version := v.Name
	return []interface{}{nameForVersion(version), v.NewTermCount, v.DeletedTermCount}
}

func rowsForVocabulary(vocabulary *metrics.Vocabulary) [][]interface{} {
	rows := make([][]interface{}, 0)
	rows = append(rows, rowForLabeledWordCount("", nil))
	for _, wc := range vocabulary.Schemas {
		rows = append(rows, rowForLabeledWordCount("schema", wc))
	}
	for _, wc := range vocabulary.Properties {
		rows = append(rows, rowForLabeledWordCount("property", wc))
	}
	for _, wc := range vocabulary.Operations {
		rows = append(rows, rowForLabeledWordCount("operation", wc))
	}
	for _, wc := range vocabulary.Parameters {
		rows = append(rows, rowForLabeledWordCount("parameter", wc))
	}
	return rows
}
