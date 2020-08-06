package main

import (
	"fmt"
	"os"

	corpus "github.com/apigee/registry/examples/corpus/corpus"
)

func main() {
	path := os.Args[1]
	c, err := corpus.ReadCorpus(path)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(-1)
	}
	c.BuildIndex()
	c.ExportOperations()
	c.ExportSchemas()
	c.ExportFields()
	c.ExportAsJSON()
}
