package models

import (
	"fmt"
	"regexp"
)

const nameRegex = "([a-zA-Z0-9-_\\.]+)"

func validateID(id string) error {
	r := regexp.MustCompile("^" + nameRegex + "$")
	m := r.FindAllStringSubmatch(id, -1)
	if m == nil {
		return fmt.Errorf("invalid id '%s'", id)
	}
	return nil
}
