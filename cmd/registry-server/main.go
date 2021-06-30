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
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/apigee/registry/server"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

func main() {
	var configPath string
	pflag.StringVarP(&configPath, "config", "c", "", "specify a configuration file")
	pflag.Parse()

	var config server.Config
	if configPath != "" {
		if err := parseConfig(&config, configPath); err != nil {
			log.Fatalf("Failed to parse config: %s", err)
		} else if err := validateConfig(config); err != nil {
			log.Fatalf("Invalid config: %s", err)
		}
	}

	addr := &net.TCPAddr{Port: 8080}
	if v, ok := os.LookupEnv("PORT"); ok {
		if port, err := strconv.Atoi(v); err != nil {
			log.Fatalf("Invalid $PORT %q: must be an integer", v)
		} else {
			addr.Port = port
		}
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %s", err)
	}
	defer listener.Close()

	srv := server.New(config)
	go srv.Start(context.Background(), listener)
	log.Printf("Listening on %s", listener.Addr())

	// Wait for an interruption signal.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	<-done
}

func parseConfig(config *server.Config, filepath string) error {
	fi, err := os.Lstat(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file info: %s", err)
	}

	// Follow symbolic links to a readable config file if applicable.
	if (fi.Mode() & os.ModeSymlink) != 0 {
		target, err := os.Readlink(filepath)
		if err != nil {
			return fmt.Errorf("failed to read symbolic link %q: %s", filepath, err)
		}
		filepath = target
	}

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %s", err)
	}

	b = []byte(os.ExpandEnv(string(b)))
	if err := yaml.Unmarshal(b, &config); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %s", err)
	}

	return nil
}

func validateConfig(c server.Config) error {
	switch c.Database {
	case "sqlite3", "postgres", "cloudsqlpostgres":
	default:
		return fmt.Errorf("invalid database value %q: must be one of [sqlite3, postgres, cloudsqlpostgres]", c.Database)
	}

	switch c.Log {
	case "fatal", "error", "warn", "info", "debug":
	default:
		return fmt.Errorf("invalid log value %q: must be one of [fatal, error, warn, info, debug]", c.Log)
	}

	if c.DBConfig == "" {
		return fmt.Errorf("invalid dbconfig %q: must not be empty", c.DBConfig)
	}

	if c.Notify && c.ProjectID == "" {
		return fmt.Errorf("invalid project %q: notifications cannot be enabled without GCP project ID", c.ProjectID)
	}

	return nil
}
