package main

import (
	"fmt"
	"os"

	corpus "github.com/apigee/registry/examples/corpus/corpus"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: corpus <path>\n")
		os.Exit(0)
	}
	path := os.Args[1]
	c, err := corpus.ReadCorpus(path)
	check(err)
	c.BuildIndex()
	c.RemoveRequestAndResponseSchemas()
	//c.ExportOperations()
	//c.ExportSchemas()
	//c.ExportFields()
	//c.ExportAsJSON()
	err = c.ExportToSheet()
	check(err)
}

func check(err error) {
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(-1)
	}
}
