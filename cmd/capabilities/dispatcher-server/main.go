package main

import (
        "log"
        "context"
        "github.com/apigee/registry/cmd/capabilities/dispatcher-server/dispatcher")

func main() {
    log.Print("Starting subscriber...")
    ctx := context.Background()

    // Setup and start the dispatcher server
    dispatcher := &dispatcher.Dispatcher{}

    if err := dispatcher.StartServer(ctx); err != nil {
        log.Printf(err.Error())
    }
    return
}