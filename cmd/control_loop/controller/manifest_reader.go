package controller

import (
	"io/ioutil"
	"fmt"
	"gopkg.in/yaml.v2"
)

type Dependency struct {
	Source string `yaml:"source"`
	Filter string `yaml:"filter"`
}

type ManifestEntry struct {
	Resource string `yaml:"resource"`
	Filter string `yaml:"filter"`
	Dependencies []Dependency `yaml:"dependencies"`
	Action string `yaml:"action"`
}

type Manifest struct {
	Project string `yaml:"project"`
	Entries []ManifestEntry `yaml:"manifest"`
}

// TODO: Add validation for pattern values and actions
func ReadManifest(filename string) (*Manifest, error) {

	buf, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    m := &Manifest{}
    err = yaml.Unmarshal(buf, m)
    if err != nil {
        return nil, fmt.Errorf("in file %q: %v", filename, err)
    }

    return m, nil
}
 