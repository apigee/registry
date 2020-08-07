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

package corpus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// ExportToSheet ...
func (corpus *Corpus) ExportToSheet() error {
	sheetid := os.Getenv("CORPUS_SHEET_ID")
	if sheetid == "" {
		return errors.New("CORPUS_SHEET_ID environment variable must be set")
	}
	ssc, err := newStatusSheetConnection(sheetid)
	if err != nil {
		return err
	}
	return ssc.updateWithCorpus(corpus)
}

type statusSheetConnection struct {
	spreadsheetID string
	sheetsService *sheets.Service
}

func newStatusSheetConnection(id string) (*statusSheetConnection, error) {
	ssc := &statusSheetConnection{spreadsheetID: id}
	usr, _ := user.Current()
	dir := usr.HomeDir
	b, err := ioutil.ReadFile(filepath.Join(dir, ".credentials/corpus.json"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	httpClient := getClient(config)
	srv, err := sheets.New(httpClient)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	ssc.sheetsService = srv
	if err != nil {
		return nil, err
	}
	return ssc, err
}

func (ssc *statusSheetConnection) fetchSheet(cellRange string) (*sheets.ValueRange, error) {
	ctx := context.Background()
	return ssc.sheetsService.Spreadsheets.Values.Get(
		ssc.spreadsheetID,
		cellRange).
		ValueRenderOption("FORMATTED_VALUE").
		DateTimeRenderOption("SERIAL_NUMBER").
		MajorDimension("ROWS").
		Context(ctx).
		Do()
}

func (ssc *statusSheetConnection) updateRows(cellRange string, values [][]interface{}) {
	ctx := context.Background()
	valueRange := &sheets.ValueRange{
		Range:          cellRange,
		MajorDimension: "ROWS",
		Values:         values,
	}
	_, err := ssc.sheetsService.Spreadsheets.Values.Update(
		ssc.spreadsheetID, valueRange.Range, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()
	if err != nil {
		log.Fatal(err)
	}
}

func (ssc *statusSheetConnection) updateWithCorpus(corpus *Corpus) error {
	// get current sheet contents
	contents, _ := ssc.fetchSheet("operations!A:E")
	log.Printf("%+v", contents)

	// update operations
	ssc.updateRows("operations!A1:E1",
		[]([]interface{}){[]interface{}{"rpc", "service", "verb", "path", "file"}},
	)
	operationRows := make([][]interface{}, 0)
	for _, op := range corpus.Operations {
		operationRows = append(operationRows, rowForOperation(op))
	}
	ssc.updateRows(fmt.Sprintf("operations!A2:E%d", 1+len(operationRows)), operationRows)

	// update schemas
	ssc.updateRows("schemas!A1:D1",
		[]([]interface{}){[]interface{}{"message", "resource name", "type", "file"}},
	)
	schemaRows := make([][]interface{}, 0)
	for _, s := range corpus.Schemas {
		schemaRows = append(schemaRows, rowForSchema(s))
	}
	ssc.updateRows(fmt.Sprintf("schemas!A2:D%d", 1+len(schemaRows)), schemaRows)

	// update fields
	ssc.updateRows("fields!A1:C1",
		[]([]interface{}){[]interface{}{"field", "message", "file"}},
	)
	fieldRows := make([][]interface{}, 0)
	for _, f := range corpus.Fields {
		fieldRows = append(fieldRows, rowForField(f))
	}
	ssc.updateRows(fmt.Sprintf("fields!A2:C%d", 1+len(fieldRows)), fieldRows)

	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file corpus-token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	usr, _ := user.Current()
	dir := usr.HomeDir
	tokFile := filepath.Join(dir, ".credentials/corpus-token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func rowForOperation(op *Operation) []interface{} {
	row := make([]interface{}, 0)
	row = append(row, op.OperationName)
	row = append(row, op.ServiceName)
	row = append(row, op.Verb)
	row = append(row, op.Path)
	row = append(row, op.FileName)
	return row
}

func rowForSchema(s *Schema) []interface{} {
	row := make([]interface{}, 0)
	row = append(row, s.SchemaName)
	row = append(row, s.ResourceName)
	row = append(row, s.ResourceType)
	row = append(row, s.FileName)
	return row
}

func rowForField(f *Field) []interface{} {
	row := make([]interface{}, 0)
	row = append(row, f.FieldName)
	row = append(row, f.SchemaName)
	row = append(row, f.FileName)
	return row
}
