// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"cloud.google.com/go/compute/metadata"
)

type WorkerRequest struct {
	Resource string
	Command  string
}

func getAuthToken() (string, error) {
	serviceURL := "http://" + os.Getenv("APG_REGISTRY_ADDRESS")
	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)
	idToken, err := metadata.Get(tokenURL)
	if err != nil {
		log.Printf("metadata.Get: failed to query id_token: %+v", err)
		return "", err
	}

	return idToken, nil
}

func RequestHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := WorkerRequest{}
	if err = json.Unmarshal(body, &req); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received Request %s", body)

	log.Print("Getting auth token...")
	idToken, err := getAuthToken()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	os.Setenv("APG_REGISTRY_TOKEN", idToken)

	split_cmd := strings.Split(req.Command, " ")
	args := append(split_cmd[1:], req.Resource)
	cmd := exec.Command(split_cmd[0], args...)
	var output []byte
	output, err = cmd.CombinedOutput()
	log.Print(string(output))
	if err != nil {
		log.Printf("Error executing command: %v", err)
		_, _ = w.Write([]byte("Execution Completed"))
		return
	}

	log.Printf("Execution Completed: \n command: %s \nresource %s", req.Command, req.Resource)
	_, _ = w.Write([]byte("Execution Completed"))
}
