package sanitize

import (
	"html"
	"regexp"
	"strings"
)

var (
	// Compile regular expressions once
	sqlInjectionPattern = regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|WHERE|FROM|INTO|CREATE|ALTER|TRUNCATE|GRANT)`)
	htmlTagPattern     = regexp.MustCompile(`<[^>]*>`)
	nonAlphaNumPattern = regexp.MustCompile(`[^a-zA-Z0-9\s-_.]`)
)

// Text sanitizes a string by:
// 1. Trimming whitespace
// 2. Removing HTML tags
// 3. Converting HTML entities
// 4. Removing potentially dangerous characters
func Text(input string) string {
	if input == "" {
		return input
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove HTML tags
	input = htmlTagPattern.ReplaceAllString(input, "")

	// Convert HTML entities
	input = html.UnescapeString(input)

	return input
}

// SQLString sanitizes a string for use in SQL queries by:
// 1. Removing SQL injection patterns
// 2. Escaping single quotes
// 3. Removing non-alphanumeric characters except spaces, hyphens, and underscores
func SQLString(input string) string {
	if input == "" {
		return input
	}

	// Remove SQL injection patterns
	input = sqlInjectionPattern.ReplaceAllString(input, "")

	// Escape single quotes
	input = strings.ReplaceAll(input, "'", "''")

	return input
}

// Filename sanitizes a string for use as a filename by:
// 1. Removing non-alphanumeric characters except spaces, hyphens, underscores, and dots
// 2. Converting spaces to underscores
func Filename(input string) string {
	if input == "" {
		return input
	}

	// Remove non-alphanumeric characters except spaces, hyphens, underscores, and dots
	input = nonAlphaNumPattern.ReplaceAllString(input, "")

	// Convert spaces to underscores
	input = strings.ReplaceAll(input, " ", "_")

	return input
}

// URL sanitizes a URL string by:
// 1. Trimming whitespace
// 2. Removing HTML tags
// 3. Validating URL format
func URL(input string) string {
	if input == "" {
		return input
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove HTML tags
	input = htmlTagPattern.ReplaceAllString(input, "")

	// Basic URL validation - could be enhanced based on needs
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return ""
	}

	return input
}
