package util

import (
	"regexp"
	"strings"
)

var invalidNameRegexp = regexp.MustCompile(`[^a-z0-9_]`)

// NormalizeName will convert the n into an low case alphanumeric value
// and the invalid characters will be replaced by '_'
func NormalizeName(n string) string {
	return invalidNameRegexp.ReplaceAllString(strings.ToLower(n), "_")
}
