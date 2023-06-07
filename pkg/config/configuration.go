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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// Configuration is stored and loaded as yaml.
type Configuration struct {
	Registry    Registry `mapstructure:"registry"`
	TokenSource string   `mapstructure:"token-source" yaml:"token-source"` // runs in shell to generate token
}

type Registry struct {
	Address  string `mapstructure:"address" yaml:"address"`   // service address
	Insecure bool   `mapstructure:"insecure" yaml:"insecure"` // if true, connect over HTTP
	Location string `mapstructure:"location" yaml:"location"`
	Project  string `mapstructure:"project" yaml:"project"`
	Token    string `mapstructure:"token" yaml:"-"` // generated from TokenSource
}

// if a name is unqualified, attempt this namespace
const default_namespace = "registry"

// Write stores the Configuration in the passed file name.
func (c Configuration) Write(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	if err := os.MkdirAll(Directory, os.FileMode(0755)); err != nil { // rwx,rx,rx
		return err
	}
	path := filepath.Join(Directory, name)
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644)) // rw,r,r
	if err != nil {
		return err
	}
	defer out.Close()
	enc := yaml.NewEncoder(out)
	return enc.Encode(c)
}

// Validate returns a ValidationError if Config is invalid.
func (c Configuration) Validate() error {
	if c.Registry.Address == "" {
		return ValidationError{
			"registry.address", "required",
		}
	}
	return nil
}

// ValidateProperty returns an UnknownPropertyError if not a valid property.
func (c Configuration) ValidateProperty(k string) error {
	kNS := k
	if !strings.Contains(k, ".") && k != "token-source" {
		kNS = default_namespace + "." + kNS
	}
	for _, p := range c.Properties() {
		if p == kNS {
			return nil
		}
	}
	return UnknownPropertyError{k}
}

// Properties returns a sorted list of all valid property names.
func (c Configuration) Properties() []string {
	props := properties(c, "")
	sort.Strings(props)
	return props
}

// FlatMap returns a Configuration as a flat Map.
func (c Configuration) FlatMap() (map[string]interface{}, error) {
	m := map[string]interface{}{}
	flat := map[string]interface{}{}
	err := mapstructure.Decode(c, &m)
	flattenMap(m, flat, "")
	return flat, err
}

// Set sets a property from a qualified or default namespace name.
func (c *Configuration) Set(k string, v interface{}) error {
	if err := c.ValidateProperty(k); err != nil {
		return err
	}
	if !strings.Contains(k, ".") && k != "token-source" {
		k = default_namespace + "." + k
	}
	return c.FromMap(map[string]interface{}{
		k: v,
	})
}

// Unset removed a property by qualified or default namespace name.
func (c *Configuration) Unset(k string) error {
	if err := c.ValidateProperty(k); err != nil {
		return err
	}
	if !strings.Contains(k, ".") && k != "token-source" {
		k = default_namespace + "." + k
	}
	return c.FromMap(map[string]interface{}{
		k: "",
	})
}

// Get gets a property from a qualified or default namespace name.
func (c *Configuration) Get(k string) (interface{}, error) {
	if err := c.ValidateProperty(k); err != nil {
		return nil, err
	}
	m, err := c.FlatMap()
	if err != nil {
		return "", fmt.Errorf("cannot decode config: %v", err)
	}
	if !strings.Contains(k, ".") && k != "token-source" {
		k = default_namespace + "." + k
	}
	return m[k], nil
}

// FromMap populates a Configuration from a flat Map.
// Existing values are overridden.
func (c *Configuration) FromMap(m map[string]interface{}) error {
	m = unflattenMap(m)

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

// Resolve derived values (eg. Registry.Token from Registry.TokenSource)
func (c *Configuration) Resolve() error {
	if c.Registry.Token == "" && c.TokenSource != "" {
		shellArgs := strings.Split(c.TokenSource, " ")
		execCmd := exec.Command(shellArgs[0], shellArgs[1:]...)
		out, err := execCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error running token source: %s", string(out))
		}
		c.Registry.Token = strings.TrimSpace(string(out))
	}
	return nil
}

// recursively collect property names
func properties(c interface{}, prefix string) []string {
	var props []string
	if len(prefix) > 0 {
		prefix += "."
	}
	rv := reflect.ValueOf(c)
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		t := rt.Field(i)
		if tv, ok := t.Tag.Lookup("mapstructure"); ok {
			switch t.Type.Kind() {
			case reflect.Struct:
				f := rv.Field(i).Interface()
				props = append(props, properties(f, prefix+tv)...)
			default:
				name := strings.Split(tv, ",")[0]
				props = append(props, prefix+name)
			}
		}
	}
	return props
}

func flattenMap(src map[string]interface{}, dest map[string]interface{}, prefix string) {
	if len(prefix) > 0 {
		prefix += "."
	}
	for k, v := range src {
		switch child := v.(type) {
		case map[string]interface{}:
			flattenMap(child, dest, prefix+k)
		default:
			dest[prefix+k] = v
		}
	}
}

func unflattenMap(src map[string]interface{}) (dest map[string]interface{}) {
	dest = map[string]interface{}{}
	for k, v := range src {
		splits := strings.Split(k, ".")
		target := dest
		for _, s := range splits[:len(splits)-1] {
			newt, ok := dest[s]
			if !ok {
				newt = map[string]interface{}{}
				dest[s] = newt
			}
			target = newt.(map[string]interface{})
		}
		target[splits[len(splits)-1]] = v
	}
	return dest
}

type UnknownPropertyError struct {
	Property string
}

func (n UnknownPropertyError) Error() string {
	return fmt.Sprintf("unknown property: %q.", n.Property)
}
