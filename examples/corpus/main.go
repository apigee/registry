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

package main

import (
	"fmt"
	"log"
	"os"

	corpus "github.com/apigee/registry/examples/corpus/corpus"
	"github.com/docopt/docopt-go"
)

func main() {
	usage := `
	Usage:
		corpus help
		corpus <path> [--sheet] [--local]
		`
	arguments, err := docopt.Parse(usage, nil, false, "Corpus 1.0", false)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	path := arguments["<path>"].(string)
	c, err := corpus.ReadCorpus(path)
	check(err)
	c.BuildIndex()
	c.RemoveRequestAndResponseSchemas()
	c.FlattenPaths()
	if arguments["--sheet"].(bool) {
		err = c.ExportToSheet()
		check(err)
	}
	if arguments["--local"].(bool) {
		c.ExportOperations()
		c.ExportSchemas()
		c.ExportFields()
		c.ExportAsJSON()
	}
}

func check(err error) {
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(-1)
	}
}
