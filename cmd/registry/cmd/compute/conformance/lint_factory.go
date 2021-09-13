package conformance

import (
	"errors"
	"fmt"
)

func CreateLinter(linter_name string) (Linter, error) {
	if linter_name == "spectral" {
		return SpectralLinter{Rules: make(map[string][]string)}, nil
	} else if linter_name == "api-linter" {
		return SpectralLinter{Rules: make(map[string][]string)}, nil
	}

	return nil, errors.New(fmt.Sprintf("Unknown Linter: %s", linter_name))
}
