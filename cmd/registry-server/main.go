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
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/apigee/registry/server"
	"github.com/spf13/viper"
)

func init() {
	// Enable environment variables.
	// e.g. REGISTRY_DATABASE_DRIVER sets database.driver config value.
	viper.SetEnvPrefix("registry")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind the REGISTRY_PROJECT_IDENTIFIER environment variable.
	// For compatibility with other tools this value should be used if it's set.
	viper.BindEnv("pubsub.project", "REGISTRY_PROJECT_IDENTIFIER")

	// Enable config files.
	viper.SetConfigName("registry-server")
	viper.AddConfigPath("$HOME/.config/registry")
	viper.AddConfigPath(".")
}

func main() {
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		log.Println(err)
	} else if err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: viper.GetInt("port"),
	})
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %s", err)
	}
	defer listener.Close()

	srv := server.New(server.Config{
		Database:  viper.GetString("database.driver"),
		DBConfig:  viper.GetString("database.dsn"),
		Log:       viper.GetString("logging.level"),
		Notify:    viper.GetBool("pubsub.enabled"),
		ProjectID: viper.GetString("pubsub.project"),
	})

	go srv.Start(context.Background(), listener)
	log.Printf("Listening on %s", listener.Addr())

	// Wait for an interruption signal.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	<-done
}
