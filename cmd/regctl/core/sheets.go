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
	"context"

	"github.com/apigee/registry/cmd/regctl/googleauth"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient represents a client of the Sheets API.
type SheetsClient struct {
	sheetID string
	service *sheets.Service
}

// NewSheetsClient creates an authenticated client of the Sheets API.
func NewSheetsClient(ctx context.Context, id string) (*SheetsClient, error) {
	httpClient, err := googleauth.GetAuthenticatedClient(ctx)
	if err != nil {
		return nil, err
	}
	sheetsClient := &SheetsClient{sheetID: id}
	sheetsClient.service, err = sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	return sheetsClient, err
}

// CreateSheet creates a new spreadsheet and configures the client to point to it.
func (sc *SheetsClient) CreateSheet(title string, sheetTitles []string) (*sheets.Spreadsheet, error) {
	s := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title:      title,
			TimeZone:   "America/Los_Angeles",
			AutoRecalc: "ON_CHANGE",
		},
	}
	if len(sheetTitles) > 0 {
		sheetArray := make([]*sheets.Sheet, 0)
		for _, sheetTitle := range sheetTitles {
			sheetArray = append(sheetArray, &sheets.Sheet{Properties: &sheets.SheetProperties{Title: sheetTitle}})
		}
		s.Sheets = sheetArray
	}
	sheet, err := sc.service.Spreadsheets.Create(s).Do()
	if err != nil {
		return nil, err
	}
	sc.sheetID = sheet.SpreadsheetId
	return sheet, err
}

// Fetch gets the contents of the configured sheet.
func (sc *SheetsClient) Fetch(ctx context.Context, cellRange string) (*sheets.ValueRange, error) {
	return sc.service.Spreadsheets.Values.Get(
		sc.sheetID,
		cellRange).
		ValueRenderOption("FORMATTED_VALUE").
		DateTimeRenderOption("SERIAL_NUMBER").
		MajorDimension("ROWS").
		Context(ctx).
		Do()
}

// Update updates a range of values in the configured sheet.
func (sc *SheetsClient) Update(ctx context.Context, cellRange string, values [][]interface{}) (*sheets.UpdateValuesResponse, error) {
	valueRange := &sheets.ValueRange{
		Range:          cellRange,
		MajorDimension: "ROWS",
		Values:         values,
	}
	return sc.service.Spreadsheets.Values.Update(
		sc.sheetID, valueRange.Range, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()
}

// FormatHeaderRow freezes and bolds the top row of the configured sheet.
func (sc *SheetsClient) FormatHeaderRow(ctx context.Context, sheetId int64) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	formatRequest := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:       sheetId,
				StartRowIndex: 0,
				EndRowIndex:   1,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: &sheets.Color{
						Red:   0.9,
						Green: 0.9,
						Blue:  0.9,
					},
					HorizontalAlignment: "CENTER",
					TextFormat: &sheets.TextFormat{
						ForegroundColor: &sheets.Color{
							Red:   0,
							Green: 0,
							Blue:  0,
						},
						FontSize: 10,
						Bold:     false,
					},
					WrapStrategy: "WRAP",
				},
			},
			Fields: "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,wrapStrategy)",
		},
	}
	freezeRequest := &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				GridProperties: &sheets.GridProperties{
					FrozenRowCount: 1,
				},
				SheetId: sheetId,
			},
			Fields: "gridProperties.frozenRowCount",
		},
	}
	request := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			formatRequest,
			freezeRequest,
		},
	}
	return sc.service.Spreadsheets.BatchUpdate(sc.sheetID, request).Context(ctx).Do()
}
