package main

import (
        "log"
        "net/http"
        "os"
        "io/ioutil"
//         "fmt"
)

func main() {
    log.Print("starting server...")
    http.HandleFunc("/", requestHandler)

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

func requestHandler(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil{
        log.Printf("ioutil.ReadAll: %v", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    log.Printf("Received request %s", body)
    w.Write([]byte("Request received"))
    return
}
