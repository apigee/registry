package conformance

import (
	"fmt"
)

// CreateLinter returns a Linter object when provided the name of a linter
func CreateLinter(linter_name string) (Linter, error) {
	if linter_name == "spectral" {
		return SpectralLinter{Rules: make(map[string][]string)}, nil
	} else if linter_name == "api-linter" {
		return ApiLinter{Rules: make(map[string][]string)}, nil
	}

	return nil, fmt.Errorf("unknown linter: %s", linter_name)
}
