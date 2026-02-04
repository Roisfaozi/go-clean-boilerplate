package pkg

import (
	"regexp"
	"strings"
)

// Slugify converts a string to a slug (lowercase, dashes instead of spaces, remove special chars)
func Slugify(s string) string {
	s = strings.ToLower(s)

	// Replace spaces with dashes
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric chars (except dashes)
	reg, _ := regexp.Compile("[^a-z0-9-]+")
	s = reg.ReplaceAllString(s, "")

	// Remove multiple dashes
	reg2, _ := regexp.Compile("-+")
	s = reg2.ReplaceAllString(s, "-")

	// Trim dashes
	s = strings.Trim(s, "-")

	if s == "" {
		s = "uuid" // fallback
	}

	return s
}
