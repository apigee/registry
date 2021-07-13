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

package googleauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Retrieve a token, saves the token, then returns the authenticated client.
func GetAuthenticatedClient(ctx context.Context) (*http.Client, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	dir := usr.HomeDir
	b, err := ioutil.ReadFile(filepath.Join(dir, ".credentials/registry.json"))
	if err != nil {
		return nil, err
	}
	// Create configuration for specified scopes.
	// If you modify these scopes, delete your previously saved token.
	config, err := google.ConfigFromJSON(b,
		"https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, err
	}
	// The file registry-token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first time.
	tokenFile := filepath.Join(dir, ".credentials/registry-token.json")
	token, err := readTokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromBrowser(ctx, config)
		saveTokenToFile(tokenFile, token)
	}
	return config.Client(ctx, token), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromBrowser(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func readTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveTokenToFile(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
