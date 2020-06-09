package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"apigov.dev/registry/cmd/registry/connection"
	"apigov.dev/registry/gapic"
	"apigov.dev/registry/models"
	rpc "apigov.dev/registry/rpc"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/sheets/v4"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export properties to a Google sheet",
	Long:  `Export properties to a Google sheet`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Export called.")

		ssc, err := NewStatusSheetConnection(SHEETID)
		if err != nil {
			log.Fatal(err)
		}

		//log.Printf("Updating sheet.")
		//checkerResults := make([]CheckerResult, 0)
		//ssc.updateWithCheckerResults(checkerResults)
		//log.Printf("Done.")

		client, err := connection.NewClient()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ctx := context.TODO()

		var name, property string
		if len(args) > 0 {
			name = args[0]
		}
		if len(args) > 1 {
			property = args[1]
		}

		if m := models.ProjectRegexp().FindAllStringSubmatch(name, -1); m != nil {
			// find all matching properties for a project
			segments := m[0]
			err = ssc.exportNamedProperty(ctx, client, segments[1], "", property)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

// BEFORE RUNNING:
// ---------------
// 1. If not already done, enable the Google Sheets API
//    and check the quota for your project at
//    https://console.developers.google.com/apis/api/sheets
// 2. Install and update the Go dependencies by running `go get -u` in the
//    project directory.
// 3. Set the SHEETID to the spreadsheet you wish to update.

const SHEETID = "13nx5d6pvb1qZWcFKR8oQTREmLVwoV1ILC8CsWnD7erQ"

const SERVICE_ACCOUNT_CREDENTIALS = "/home/tim/.credentials/registrydemo.json"

type CheckerResult struct {
	api      string
	command  string
	messages []string
}

type StatusSheetConnection struct {
	spreadsheetId string
	sheetsService *sheets.Service
}

func NewStatusSheetConnection(id string) (*StatusSheetConnection, error) {
	ssc := &StatusSheetConnection{spreadsheetId: id}
	ctx := context.Background()

	// read credentials for a service account that has access to the sheet
	jsonData, err := ioutil.ReadFile(SERVICE_ACCOUNT_CREDENTIALS)
	credentials, err := google.CredentialsFromJSON(ctx, jsonData, sheets.DriveScope, sheets.DriveFileScope, sheets.SpreadsheetsScope)
	client := oauth2.NewClient(ctx, credentials.TokenSource)
	if err != nil {
		return nil, err
	}
	ssc.sheetsService, err = sheets.New(client)
	if err != nil {
		return nil, err
	}
	return ssc, err
}

func (ssc *StatusSheetConnection) fetch() (*sheets.ValueRange, error) {
	ctx := context.Background()
	return ssc.sheetsService.Spreadsheets.Values.Get(
		ssc.spreadsheetId,
		"Summary!A:A").
		ValueRenderOption("FORMATTED_VALUE").
		DateTimeRenderOption("SERIAL_NUMBER").
		MajorDimension("ROWS").
		Context(ctx).
		Do()
}

func (ssc *StatusSheetConnection) updateRows(cellRange string, values [][]interface{}) {
	ctx := context.Background()
	valueRange := &sheets.ValueRange{
		Range:          cellRange,
		MajorDimension: "ROWS",
		Values:         values,
	}
	_, err := ssc.sheetsService.Spreadsheets.Values.Update(
		ssc.spreadsheetId, valueRange.Range, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()
	if err != nil {
		log.Fatal(err)
	}
}

func matching(checkerResults []CheckerResult, api string) *CheckerResult {
	for _, checkerResult := range checkerResults {
		if checkerResult.api == api {
			return &checkerResult
		}
	}
	return nil
}

func rowForResult(checkerResult *CheckerResult, withDetails bool) []interface{} {
	row := make([]interface{}, 0)
	row = append(row, (*checkerResult).api)
	if withDetails {
		row = append(row, (*checkerResult).command)
		row = append(row, strings.Join((*checkerResult).messages, "\n"))
	} else {
		row = append(row, len((*checkerResult).messages))
	}
	return row
}

func (ssc *StatusSheetConnection) updateWithCheckerResults(checkerResults []CheckerResult) {
	// reset headings
	ssc.updateRows("Summary!A1:B1",
		[]([]interface{}){[]interface{}{"Spec", "Operation Count"}},
	)

	// get current sheet contents
	contents, _ := ssc.fetch()

	// prepare update
	summary := make([][]interface{}, 0)
	seen := make(map[string]bool, 0)

	// for each API in the sheet...
	for i, v := range contents.Values {
		if i == 0 {
			continue // skip header
		}
		// get the analysis results
		if len(v) > 0 {
			api := v[0].(string)
			checkerResult := matching(checkerResults, api)
			if checkerResult != nil {
				// if we have it, add it to the table
				summary = append(summary, rowForResult(checkerResult, false))
			} else {
				// if we don't have it, mark the api as unknown
				summary = append(summary, []interface{}{api, "unknown"})
			}
			seen[api] = true
		} else {
			// mark any weird table rows
			summary = append(summary, []interface{}{"unknown", "unknown"})
		}
	}
	// now go back through the checker results and add any that weren't in the table
	for _, checkerResult := range checkerResults {
		if !seen[checkerResult.api] {
			summary = append(summary, rowForResult(&checkerResult, false))
		}
	}
	ssc.updateRows(fmt.Sprintf("Summary!A2:B%d", 1+len(summary)), summary)
}

var values map[string]int64

func (ssc *StatusSheetConnection) exportNamedProperty(ctx context.Context, client *gapic.RegistryClient, projectID string, subject string, relation string) error {
	request := &rpc.ListPropertiesRequest{
		Parent: subject,
		Filter: fmt.Sprintf("property_id = \"%s\"", relation),
	}
	values = make(map[string]int64, 0)
	it := client.ListProperties(ctx, request)
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		values[property.Subject] = property.GetInt64Value()
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	summary := make([][]interface{}, 0)
	sort.Strings(keys)
	for _, k := range keys {
		log.Printf("%s: %d", k, values[k])

		row := make([]interface{}, 0)
		row = append(row, k)
		row = append(row, values[k])
		summary = append(summary, row)
	}
	// reset headings
	ssc.updateRows("Summary!A1:B1",
		[]([]interface{}){[]interface{}{"Spec", "Operation Count"}},
	)
	ssc.updateRows(fmt.Sprintf("Summary!A2:B%d", 1+len(summary)), summary)

	return nil
}
