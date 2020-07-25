package names

import (
	"fmt"
	"regexp"
)

// We might extend this to all characters that do not require escaping.
// See "Resource ID Segments" in https://aip.dev/122.
const NameRegex = "([a-zA-Z0-9-_\\.]+)"

// Generated revision names are lowercase hex strings, but we also
// allow user-specified revision tags which can be mixed-case strings
// containing dashes.
const RevisionRegex = "(@[a-zA-z0-9-]+)?"

func ValidateID(id string) error {
	r := regexp.MustCompile("^" + NameRegex + "$")
	m := r.FindAllStringSubmatch(id, -1)
	if m == nil {
		return fmt.Errorf("invalid id '%s'", id)
	}
	return nil
}

func ValidateRevision(s string) error {
	r := regexp.MustCompile("^" + RevisionRegex + "$")
	m := r.FindAllStringSubmatch(s, -1)
	if m == nil {
		return fmt.Errorf("invalid revision '%s'", s)
	}
	return nil
}
