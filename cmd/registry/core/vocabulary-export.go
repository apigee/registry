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

	metrics "github.com/googleapis/gnostic/metrics"
)

func ExportVocabularyToSheet(name string, vocabulary *metrics.Vocabulary) (string, error) {
	sheetsClient, err := NewSheetsClient("")
	if err != nil {
		return "", err
	}
	sheet, err := sheetsClient.CreateSheet(name, []string{"Everything", "Schemas", "Properties", "Operations", "Parameters"})
	if err != nil {
		return "", err
	}
	for _, s := range sheet.Sheets {
		_, err := sheetsClient.FormatHeaderRow(s.Properties.SheetId)
		if err != nil {
			return "", err
		}
	}
	{
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
		_, err = sheetsClient.Update(fmt.Sprintf("Everything!A1:C%d", len(rows)), rows)
		if err != nil {
			return "", err
		}
	}
	// create update function as a closure to simplify multiple uses (below)
	updateSheet := func(title string, pairs []*metrics.WordCount) error {
		rows := make([][]interface{}, 0)
		rows = append(rows, rowForWordCount(nil))
		for _, wc := range pairs {
			rows = append(rows, rowForWordCount(wc))
		}
		_, err = sheetsClient.Update(fmt.Sprintf("%s!A1:C%d", title, len(rows)), rows)
		return err
	}
	updateSheet("Schemas", vocabulary.Schemas)
	updateSheet("Properties", vocabulary.Properties)
	updateSheet("Operations", vocabulary.Operations)
	updateSheet("Parameters", vocabulary.Parameters)
	return sheet.SpreadsheetUrl, nil
}

func rowForLabeledWordCount(kind string, wc *metrics.WordCount) []interface{} {
	if wc == nil {
		return []interface{}{"kind", "word", "count"}
	}
	return []interface{}{kind, wc.Word, wc.Count}
}

func rowForWordCount(wc *metrics.WordCount) []interface{} {
	if wc == nil {
		return []interface{}{"word", "count"}
	}
	return []interface{}{wc.Word, wc.Count}
}
