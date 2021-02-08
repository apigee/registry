package linterapp

import (
        "context"
        "fmt"
        "log"
        "net/http"
        "os"
        "io/ioutil"
        "strings"
        "encoding/json"
        "github.com/apigee/registry/cmd/registry/cmd"
        "github.com/apigee/registry/rpc"
	    "github.com/apigee/registry/connection"
        "cloud.google.com/go/compute/metadata"
	    "github.com/golang/protobuf/jsonpb"
)

type PubSubMessage struct {
        Message struct {
                Data []byte `json:"data,omitempty"`
                ID   string `json:"id"`
        } `json:"message"`
        Subscription string `json:"subscription"`
}

func getAuthToken() (string, error) {
        serviceURL := "https://" + strings.Split(os.Getenv("APG_REGISTRY_ADDRESS"), ":")[0]
        tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)
        idToken, err := metadata.Get(tokenURL)
        if err != nil {
                log.Printf("metadata.Get: failed to query id_token: %+v", err)
                return "", err
        }

        return idToken, nil
}

func readPubsubMessage(w http.ResponseWriter, r *http.Request) string {
        var m PubSubMessage
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
                log.Printf("ioutil.ReadAll: %v", err)
                http.Error(w, "Bad Request", http.StatusBadRequest)
                return ""
        }
        if err := json.Unmarshal(body, &m); err != nil {
                log.Printf("json.Unmarshal: %v", err)
                http.Error(w, "Bad Request", http.StatusBadRequest)
                return ""
        }

        data := string(m.Message.Data)
        log.Printf("%s", data)
        return data
}

func RequestHandler(w http.ResponseWriter, r *http.Request) {

    data := readPubsubMessage(w, r)
    if data == "" {
        return
    }
    message := rpc.Notification{}
    if err := jsonpb.UnmarshalString(data, &message); err != nil {
        log.Printf("json.Unmarshal: %v", err)
        fmt.Fprintf(w, "Wrong message format")
        return
    }

    if message.Change != rpc.Notification_DELETED {

        log.Print("Getting oauth token")
        idToken, err := getAuthToken()
        if err != nil {
            log.Print(err.Error())
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        os.Setenv("APG_REGISTRY_TOKEN", idToken)

        specName := strings.Split(message.Resource, "@")[0]
        ctx := context.TODO()
        log.Print("Creating connection...")
        client, err := connection.NewClient(ctx)
        if err != nil {
            log.Print(err.Error())
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        log.Print("Done")

        lint_task := &cmd.ComputeLintTask{
        Ctx: ctx,
        Client: client,
        SpecName: specName,
        Linter: "",
        }

        err = lint_task.Run()
        if err != nil {
            log.Print(err.Error())
            fmt.Fprintf(w, "Error computing lint")
            return
        }
    }
    fmt.Fprintf(w, "Done")
    return
}