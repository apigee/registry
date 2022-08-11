// Copyright 2022 Google LLC. All Rights Reserved.
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
	"github.com/apigee/registry/pkg/config"
)

// Config configures the client.
type Config struct {
	Address  string `mapstructure:"address"`  // service address
	Insecure bool   `mapstructure:"insecure"` // if true, connect over HTTP
	Token    string `mapstructure:"token"`    // bearer token
}

// ActiveConfig returns the active config.
func ActiveConfig() (Config, error) {
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

	reg := c.Registry
	config := Config{
		Address:  reg.Address,
		Insecure: reg.Insecure,
		Token:    reg.Token,
	}

	return config, err
}
