package core

import (
	"errors"
	"fmt"
)

func CreateLinter(linter_name string) (Linter, error) {
	if linter_name == "spectral" {
		return SpectralLinter{Name: "spectral", Rules: make(map[string][]string)}, nil
	}

	return nil, errors.New(fmt.Sprintf("Unknown Linter: %s", linter_name))
}
