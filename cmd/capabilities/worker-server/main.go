package main

import (
        "log"
        "net/http"
        "os"
        "github.com/apigee/registry/cmd/capabilities/worker-server/worker"
)

func main() {
    log.Print("starting server...")
    http.HandleFunc("/", worker.RequestHandler)

    // Determine port for HTTP service.
    port := os.Getenv("PORT")
    if port == "" {
            port = "8080"
            log.Printf("defaulting to port %s", port)
    }

    // Start HTTP server.
    log.Printf("listening on port %s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
            log.Fatal(err)
    }
}