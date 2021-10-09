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
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/apex/log"
	"github.com/apigee/registry/servers/registry"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

// ServerConfig is the top-level configuration structure.
type ServerConfig struct {
	// Server port. If unset or zero, an open port will be assigned.
	Port     int            `yaml:"port"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
	Pubsub   PubsubConfig   `yaml:"pubsub"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	// Driver for the database connection.
	// Values: [ sqlite3, postgres, cloudsqlpostgres ]
	Driver string `yaml:"driver"`
	// Config for the database connection. The format is a data source name (DSN).
	// PostgreSQL Reference: See "Connection Strings" at https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
	// SQLite Reference: See "URI filename examples" at https://www.sqlite.org/c3ref/open.html
	Config string `yaml:"config"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	// Level of logging to print to standard output.
	// Values: [ debug, info, warn, error, fatal ]
	Level string `yaml:"level"`
	// Format of log entries.
	// Options: [ json, text ]
	Format string `yaml:"format"`
}

// PubsubConfig holds pubsub (notification) configuration.
type PubsubConfig struct {
	// Enable Pub/Sub for event notification publishing.
	// Values: [ true, false ]
	Enable bool `yaml:"enable"`
	// Project ID of the Google Cloud project to use for Pub/Sub.
	// Reference: https://cloud.google.com/resource-manager/docs/creating-managing-projects
	Project string `yaml:"project"`
}

// default configuration
var config = ServerConfig{
	Port: 8080,
	Database: DatabaseConfig{
		Driver: "sqlite3",
		Config: "file:/tmp/registry.db",
	},
	Logging: LoggingConfig{
		Level:  "info",
		Format: "text",
	},
	Pubsub: PubsubConfig{
		Enable:  false,
		Project: "",
	},
}

func main() {
	var configPath string
	pflag.StringVarP(&configPath, "configuration", "c", "", "The server configuration file to load.")
	pflag.Parse()

	if configPath != "" {
		log.Infof("Loading configuration from %s", configPath)
		raw, err := ioutil.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Failed to open config file: %s", err)
		}
		// expand environment variables before unmarshaling
		expanded := []byte(os.ExpandEnv(string(raw)))
		err = yaml.Unmarshal(expanded, &config)
		if err != nil {
			log.Fatalf("Failed to read config file: %s", err)
		}
	}

	if err := validateConfig(); err != nil {
		log.Fatalf("Invalid configuration: %s", err)
	}

	log.Infof("Configured port %d", config.Port)
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: config.Port,
	})
	if err != nil {
		log.Fatalf("Failed to create TCP listener: %s", err)
	}
	defer listener.Close()

	srv := registry.New(registry.Config{
		Database:  config.Database.Driver,
		DBConfig:  config.Database.Config,
		LogLevel:  config.Logging.Level,
		LogFormat: config.Logging.Format,
		Notify:    config.Pubsub.Enable,
		ProjectID: config.Pubsub.Project,
	})

	go srv.Start(context.Background(), listener)
	log.Infof("Listening on %s", listener.Addr())

	// Wait for an interruption signal.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	<-done
}

func validateConfig() error {
	if config.Port < 0 {
		return fmt.Errorf("invalid port %q: must be non-negative", config.Port)
	}

	switch driver := config.Database.Driver; driver {
	case "sqlite3", "postgres", "cloudsqlpostgres":
	default:
		return fmt.Errorf("invalid database.driver %q: must be one of [sqlite3, postgres, cloudsqlpostgres]", driver)
	}

	switch level := config.Logging.Level; level {
	case "fatal", "error", "warn", "info", "debug":
	default:
		return fmt.Errorf("invalid logging.level %q: must be one of [fatal, error, warn, info, debug]", level)
	}

	switch format := config.Logging.Format; format {
	case "json", "text":
	default:
		return fmt.Errorf("invalid logging format %q: must be one of [json, text]", format)
	}

	if project := config.Pubsub.Project; config.Pubsub.Enable && project == "" {
		return fmt.Errorf("invalid pubsub.project %q: pubsub cannot be enabled without GCP project ID", project)
	}

	return nil
}
