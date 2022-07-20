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
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
)

// Flags defines Flags that may be bound to a Configuration. Use like:
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
	Flags.StringP("config", "c", "", "Name of a configuration profile or path to config file")
	Flags.String("registry.address", "", "the server and port of the registry api (eg. localhost:8080)")
	Flags.Bool("registry.insecure", false, "if specified, client connects via http (not https)")
	Flags.String("registry.token", "", "the token to use for authorization to registry")
}

// Config configure the client.
type Config struct {
	Address  string `mapstructure:"address"`  // service address
	Insecure bool   `mapstructure:"insecure"` // if true, connect over HTTP
	Token    string `mapstructure:"token"`    // bearer token
}

// Validate returns an error if Config is invalid.
func (c Config) Validate() error {
	if c.Address == "" {
		return ValidationError{
			"registry.address", "required",
		}
	}
	return nil
}

// Write stores the Config in the passed file name
// within the configpath.
func (c Config) Write(name string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("registry.Address", c.Address)
	v.Set("registry.Insecure", c.Insecure)
	path := filepath.Join(configPath, name)
	return v.WriteConfigAs(path)
}

// Names returns the setting names in a Config.
func (c Config) Names() []string {
	var names []string
	rt := reflect.TypeOf(c)
	for i := 0; i < rt.NumField(); i++ {
		t := rt.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		names = append(names, strings.Split(tv, ",")[0])
	}
	return names
}

// AsMap returns the Config as a Map.
func (c Config) AsMap() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := mapstructure.Decode(c, &m)
	return m, err
}

func (c *Config) FromMap(m map[string]interface{}) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           c,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}

func AllConfigs() (map[string]Config, error) {
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		return nil, err
	}

	var errors error
	allConfigs := make(map[string]Config)
	for _, file := range files {
		if !file.IsDir() {
			s, err := ReadConfig(file.Name())
			if err != nil {
				errors = multierr.Append(errors, err)
				continue
			}
			allConfigs[file.Name()] = s
		}
	}

	return allConfigs, nil
}

// ActiveConfig determines the active configuration name,
// reads the configuration, validates it, and returns
// a Config and possibly an error.
// If `config` flag exists, overrides active_config file.
func ActiveConfig() (Config, error) {
	var err error
	name, _ := Flags.GetString("config")
	if name == "" {
		name, err = ActiveConfigName()
		if err != nil {
			return Config{}, err
		}
	}

	return ReadValidConfig(name)
}

// ActivateConfig sets the active configuration file.
// Will error if config doesn't exist.
func ActivateConfig(name string) error {
	_, err := ReadConfig(name)
	if err != nil {
		return err
	}

	f := filepath.Join(configPath, activeConfigPointerFilename)
	return ioutil.WriteFile(f, []byte(name), os.FileMode(0644)) // rw,r,r
}

// ReadValidConfig reads the specified configuration
// and validates it. An error is returned if the Config could not
// be read or is not valid. Binds to standard Flags().
func ReadValidConfig(name string) (config Config, err error) {
	config, err = ReadConfig(name)
	if err != nil {
		return
	}
	err = config.Validate()
	return config, err
}

// ReadConfig loads Config from yaml file matching `name`. If name
// contains a path, the file will be read from that path, otherwise
// the path is assumed as: ~/.config/registry.
// Setting values are prioritized in order from: flags, env vars, file.
// If name is empty, no file will be loaded and only flags and env vars
// will be used.
// Client can call Flags() to get the standard list.
// env vars: APG_REGISTRY_ADDRESS, APG_REGISTRY_INSECURE, APG_REGISTRY_TOKEN
// flag names: registry.address, registry.insecure, registry.token
func ReadConfig(name string) (config Config, err error) {
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
		Config Config `mapstructure:"registry"`
	}{}
	if err = v.Unmarshal(&reg); err != nil {
		return
	}
	config = reg.Config
	return
}

// DeleteConfig deletes a configuration.
// Will error if active or missing (*os.PathError).
func DeleteConfig(name string) error {
	active, err := ActiveConfigName()
	if err != nil {
		return err
	}
	if name == active {
		return fmt.Errorf("Cannot delete active configuration")
	}

	f := filepath.Join(configPath, activeConfigPointerFilename)
	return os.Remove(f)
}

// returns the config file to use from ~/.config/active_config.
// Returns "" if active_config is not found.
func ActiveConfigName() (string, error) {
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
