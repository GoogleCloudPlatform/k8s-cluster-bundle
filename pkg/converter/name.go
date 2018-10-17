package converter

import (
	"regexp"
	"strings"
)

var firstCharRegexp = regexp.MustCompile(`^[^a-z0-9]`)
var nameRegexp = regexp.MustCompile(`[^a-z0-9_.-]`)
var lastCharRegexp = regexp.MustCompile(`[^a-z0-9]$`)

// SanitizeName sanitizes a metadata.name field, replacing unsafe characters
// with _ and truncating if it's longer than 253 characters.
func SanitizeName(name string) string {
	name = strings.ToLower(name)
	name = nameRegexp.ReplaceAllString(name, "_")
	name = firstCharRegexp.ReplaceAllString(name, "z")
	name = lastCharRegexp.ReplaceAllString(name, "z")
	if len(name) >= 254 {
		name = name[0:253]
	}
	return name
}
