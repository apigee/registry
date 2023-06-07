// Copyright 2022 Google LLC.
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

package connection

import (
	"errors"
	"path"
	"strings"

	"github.com/apigee/registry/pkg/config"
)

// Config configures the client.
type Config struct {
	Address  string `mapstructure:"address"`  // service address
	Insecure bool   `mapstructure:"insecure"` // if true, connect over HTTP
	Location string `mapstructure:"location"` // optional
	Project  string `mapstructure:"project"`  // optional
	Token    string `mapstructure:"token"`    // bearer token
}

// If set, ActiveConfig() returns this configuration.
// This is intended for use in testing.
var active *Config

// SetConfig forces the active configuration to use the specified value.
func SetConfig(config Config) {
	active = &config
}

// ActiveConfig returns the active config.
func ActiveConfig() (Config, error) {
	if active != nil {
		return *active, nil
	}

	name, err := config.ActiveName()
	if err != nil {
		return Config{}, err
	}

	return ReadConfig(name)
}

// Reads a Config from a file. If name is empty, no
// file will be loaded and only bound flags and
// env vars will be used.
func ReadConfig(name string) (Config, error) {
	c, err := config.ReadValid(name)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Address:  c.Registry.Address,
		Insecure: c.Registry.Insecure,
		Location: c.Registry.Location,
		Project:  c.Registry.Project,
		Token:    c.Registry.Token,
	}

	return config, err
}

// FQName ensures the project and location, if available,
// are properly included to make ensure the resource name
// is fully qualified.
func (c Config) FQName(name string) string {
	name = strings.TrimPrefix(name, "/")
	if !strings.HasPrefix(name, "projects") && c.Project != "" {
		if strings.HasPrefix(name, "locations") {
			name = path.Join("projects", c.Project, name)
		} else if c.Location != "" {
			name = path.Join("projects", c.Project, "locations", c.Location, name)
		} else {
			// Use "global" as the default location.
			name = path.Join("projects", c.Project, "locations", "global", name)
		}
	}
	return name
}

func (c Config) ProjectWithLocation() (string, error) {
	if c.Project == "" {
		return "", errors.New("registry.project is not specified")
	}
	if c.Location == "" {
		return "projects/" + c.Project + "/locations/global", nil
	}
	return "projects/" + c.Project + "/locations/" + c.Location, nil
}
