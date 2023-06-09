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

package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
)

const ActivePointerFilename = "active_config"

var (
	ErrNoActiveConfiguration = fmt.Errorf("no active configuration")

	// Flags defines Flags that may be bound to a Configuration. Use like:
	// `cmd.PersistentFlags().AddFlagSet(connection.Flags)`
	Flags *pflag.FlagSet = CreateFlagSet()

	// Directory is $HOME/config/registry
	Directory             string
	ErrCannotDeleteActive = fmt.Errorf("cannot delete active configuration")
	ErrReservedConfigName = fmt.Errorf("%q is reserved", ActivePointerFilename)

	envBindings    = []string{"registry.address", "registry.insecure", "registry.token"}
	envKeyReplacer = strings.NewReplacer(".", "_")
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	Directory = filepath.Join(home, ".config/registry")
}

func CreateFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("registry", pflag.ExitOnError)
	flags.StringP("config", "c", "", "name of a configuration profile or path to config file")
	flags.String("registry.address", "", "the server and port of the Registry API (eg. localhost:8080)")
	flags.String("address", "", "the server and port of the Registry API (eg. localhost:8080)")
	flags.Bool("registry.insecure", false, "if specified, client connects via http (not https)")
	flags.String("registry.location", "", "the API Registry location")
	flags.String("registry.project", "", "the API Registry project")
	flags.String("registry.token", "", "the token to use for authorization to the API Registry")
	return flags
}

// Configurations returns stored Configurations by name
func Configurations() (map[string]Configuration, error) {
	files, err := os.ReadDir(Directory)
	if err != nil {
		return nil, err
	}

	var errors error
	configs := make(map[string]Configuration)
	for _, file := range files {
		if !file.IsDir() && file.Name() != ActivePointerFilename {
			s, err := Read(file.Name())
			if err != nil {
				errors = multierr.Append(errors, err)
				continue
			}
			configs[file.Name()] = s
		}
	}
	if errors != nil {
		return nil, errors
	}

	return configs, nil
}

// ValidateName ensures a Configuration name is valid
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if name == ActivePointerFilename {
		return ErrReservedConfigName
	}

	if dir, _ := filepath.Split(name); dir != "" {
		return fmt.Errorf("%q must not include a path", name)
	}
	return nil
}

// Active determines the active Configuration name,
// reads the Configuration, validates it, and returns
// a Configuration and possibly an error.
// If `config` flag exists, overrides active_config file.
func Active() (c Configuration, err error) {
	name, _ := Flags.GetString("config")
	if name == "" {
		name, err = ActiveName()
		if err != nil {
			return Configuration{}, err
		}
	}

	return ReadValid(name)
}

// ActiveRaw reads the active file without env or flag bindings.
func ActiveRaw() (name string, c Configuration, err error) {
	name, _ = Flags.GetString("config")
	if name == "" {
		name, err = ActiveName()
		if err != nil {
			return name, Configuration{}, err
		}
	}
	if name == "" {
		return name, Configuration{}, ErrNoActiveConfiguration
	}

	c, err = Read(name)
	return
}

// Activate sets the active Configuration file.
// Will error if file doesn't exist.
func Activate(name string) error {
	_, err := Read(name)
	if err != nil {
		return err
	}

	f := filepath.Join(Directory, ActivePointerFilename)
	return os.WriteFile(f, []byte(name), os.FileMode(0644)) // rw,r,r
}

// ReadValid reads the specified Configuration, resolves it, and
// validates it. If name is empty, no file will be loaded and only
// bound flags and env vars will be used. An error is returned if
// the Configuration could not be read from a non-empty name, is not
// cannot be resolved, or is not valid once loaded and resolved.
// Binds to standard Flags and env vars. If name is empty, no file
// will be loaded and only bound flags and env vars will be used.
// See CreateFlagSet() and bindEnvs().
func ReadValid(name string) (c Configuration, err error) {
	dir, file := filepath.Split(name)
	var r io.Reader = &bytes.Buffer{}
	if file != "" {
		if dir == "" {
			name = filepath.Join(Directory, file)
		}
		r, err = os.Open(name)
		if err != nil {
			return
		}
		defer r.(*os.File).Close()
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err = v.BindPFlags(Flags); err != nil {
		return
	}
	if err = bindEnvs(v); err != nil {
		return
	}
	if err = v.ReadConfig(r); err != nil {
		return
	}
	if err = v.Unmarshal(&c); err != nil {
		return
	}
	if err = c.Resolve(); err != nil {
		return
	}
	if err = c.Validate(); err != nil {
		return
	}

	return
}

// Read loads a Configuration from yaml file matching `name`. If name
// contains a path, the file will be read from that path, otherwise
// the path is assumed as: ~/.config/registry. Does a simple read from the
// file: does not bind to env vars or flags, resolve, or validate.
// See also: ReadValid()
func Read(name string) (c Configuration, err error) {
	if err = ValidateName(name); err != nil {
		return
	}

	dir, file := filepath.Split(name)
	if dir == "" {
		name = filepath.Join(Directory, file)
	}
	var r io.Reader
	if r, err = os.Open(name); err != nil {
		return
	}
	defer r.(*os.File).Close()

	v := viper.New()
	v.SetConfigType("yaml")
	if err = v.ReadConfig(r); err != nil {
		return
	}
	if err = v.Unmarshal(&c); err != nil {
		return
	}
	return c, err
}

// Delete deletes a Configuration.
// Will error if active or missing (*os.PathError).
func Delete(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	active, err := ActiveName()
	if err != nil {
		return err
	}
	if name == active {
		return ErrCannotDeleteActive
	}

	f := filepath.Join(Directory, name)
	return os.Remove(f)
}

// Returns the active file from ~/.config/active_config.
// Returns "" if active_config is not found.
func ActiveName() (string, error) {
	f := filepath.Join(Directory, ActivePointerFilename)
	bytes, err := os.ReadFile(f)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), err
}

// Binds environment vars to populate config
func bindEnvs(v *viper.Viper) error {
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(envKeyReplacer)
	for _, env := range envBindings {
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
