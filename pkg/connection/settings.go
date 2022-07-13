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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Flags defines Flags that may be bound to the Settings. Use like:
// `cmd.PersistentFlags().AddFlagSet(connection.Flags)`
var Flags *pflag.FlagSet

var configPath string
var activeConfigPointerFilename = "active_config"

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	configPath = filepath.Join(home, ".config/registry")

	Flags = pflag.NewFlagSet("registry", pflag.ExitOnError)
	Flags.StringP("config", "c", "", "Name of a settings profile or path to config file")
	Flags.String("registry.address", "", "the server and port of the registry api (eg. localhost:8080)")
	Flags.Bool("registry.insecure", false, "set to true if client must connect via http (not https)")
	Flags.String("registry.token", "", "the token to use for authorization to registry")
}

// Settings configure the client.
type Settings struct {
	Address  string `mapstructure:"address"`  // service address
	Insecure bool   `mapstructure:"insecure"` // if true, connect over HTTP
	Token    string `mapstructure:"token"`    // bearer token
}

func (s Settings) Validate() error {
	if s.Address == "" {
		return ValidationError{
			"registry.address", "required",
		}
	}
	return nil
}

func (s Settings) Write(name string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("registry.Address", s.Address)
	v.Set("registry.Insecure", s.Insecure)
	path := filepath.Join(configPath, name)
	return v.WriteConfigAs(path)
}

// ActiveSettings determines the active configuration name,
// reads the configuration, validates it, and returns
// a valid Settings and possibly an error.
// If `config` flag exists, overrides active_config file.
func ActiveSettings() (Settings, error) {
	var err error
	name, _ := Flags.GetString("config")
	if name == "" {
		name, err = readActiveConfig()
		if err != nil {
			return Settings{}, err
		}
	}

	return ReadValidSettings(name)
}

// SetActiveConfigFile sets the active configuration file.
// name is normally a file name with no path.
// will not error if config file doesn't exist
func SetActiveConfigFile(name string) error {
	f := filepath.Join(configPath, activeConfigPointerFilename)
	return ioutil.WriteFile(f, []byte(name), os.FileMode(0644)) // rw,r,r
}

// ReadValidSettings reads the specified configuration
// and validates it. An error is returned if the Settings could not
// be read or are not valid. Binds to standard Flags().
func ReadValidSettings(name string) (settings Settings, err error) {
	settings, err = ReadSettings(name)
	if err != nil {
		return
	}
	err = settings.Validate()
	return settings, err
}

// ReadSettings loads Settings from yaml file matching `name`. If name
// contains a path, the file will be read from that path, otherwise
// the path is assumed as: ~/.config/registry.
// Setting values are prioritized in order from: flags, env vars, file.
// If name is empty, no file will be loaded and only flags and env vars
// will be used.
// Client can call Flags() to get the standard list.
// env vars: APG_REGISTRY_ADDRESS, APG_REGISTRY_INSECURE, APG_REGISTRY_TOKEN
// flag names: registry.address, registry.insecure, registry.token
func ReadSettings(name string) (settings Settings, err error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err = v.BindPFlags(Flags); err != nil {
		return
	}
	if err = bindEnvs(v); err != nil {
		return
	}

	dir, file := filepath.Split(name)
	v.SetConfigName(file)
	if dir != "" {
		v.AddConfigPath(dir)
	} else {
		v.AddConfigPath(configPath)
	}

	if err = v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && name == "" {
			err = nil
		} else {
			return
		}
	}

	// add wrapper "registry.xxx " for unmarshal
	reg := struct {
		Settings Settings `mapstructure:"registry"`
	}{}
	if err = v.Unmarshal(&reg); err != nil {
		return
	}
	settings = reg.Settings
	return
}

// returns the config file to use from ~/.config/active_config.
// Returns "" if active_config is not found.
func readActiveConfig() (string, error) {
	f := filepath.Join(configPath, activeConfigPointerFilename)
	bytes, err := ioutil.ReadFile(f)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	return strings.TrimSpace(string(bytes)), err
}

// binds environment vars to populate config
func bindEnvs(v *viper.Viper) error {
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("APG")
	bindings := []string{"registry.address", "registry.insecure", "registry.token"}
	for _, env := range bindings {
		if err := v.BindEnv(env); err != nil {
			return err
		}
	}
	return nil
}

type ValidationError struct {
	Field      string
	Validation string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Validation)
}
