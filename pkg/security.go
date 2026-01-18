package pkg

import (
	"html"
	"regexp"
	"strings"
)

func ContainsSQLInjection(input string) bool {
	// Added \b boundaries to keywords to prevent false positives like "selection" or "updateable"
	// Kept symbols and prefixes as is
	sqlInjectionPattern := `(?i)('|--|;|/\*|\*/|xp_|sp_|\b(exec|execute|union|select|insert|update|delete|drop|alter|create|truncate|grant|revoke)\b)`
	matched, _ := regexp.MatchString(sqlInjectionPattern, input)
	return matched
}

func SanitizeString(input string) string {
	output := strings.TrimSpace(input)
	output = html.EscapeString(output)
	return output
}
