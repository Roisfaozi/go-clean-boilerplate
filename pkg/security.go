package pkg

import (
	"html"
	"regexp"
	"strings"
)

func ContainsSQLInjection(input string) bool {
	sqlInjectionPattern := `(?i)('|--|;|/\*|\*/|xp_|sp_|exec|execute|union|select|insert|update|delete|drop|alter|create|truncate|grant|revoke)`
	matched, _ := regexp.MatchString(sqlInjectionPattern, input)
	return matched
}

func SanitizeString(input string) string {
	output := strings.TrimSpace(input)
	output = html.EscapeString(output)
	return output
}
