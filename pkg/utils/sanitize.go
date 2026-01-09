package utils

import (
	"regexp"
	"strings"
)

func SanitizeFilename(s string) string {
	// Remove scheme
	s = strings.Replace(s, "https://", "", 1)
	s = strings.Replace(s, "http://", "", 1)

	// Replace non-alphanumeric with underscore
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	s = reg.ReplaceAllString(s, "_")

	return strings.Trim(s, "_")
}
