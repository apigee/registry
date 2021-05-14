package main

import (
        "log"
	    "github.com/apigee/registry/cmd/control_loop/controller"
	    "os"
)

// TODO: Convert this into a cobra command
// Cuurently execute using: "./main manifest.yaml"
func main() {
	manifestPath := os.Args[1]
	manifest, err := controller.ReadManifest(manifestPath)
	if err!=nil {
		log.Fatal(err.Error())
	}

	actions, err := controller.ProcessManifest(manifest)
	if err!=nil {
		log.Fatal(err.Error())
	}

	log.Print("Actions:")
	for i, a := range actions {
		log.Printf("%d: %s", i, a)
	}
}

