package slugkit

import (
	"regexp"
	"strings"
)

var (
	nonAlnum     = regexp.MustCompile(`[^a-z0-9]+`)
	trailingDash = regexp.MustCompile(`-+$`)
)

// Sanitize lowercases, replaces non-alphanumeric runs with hyphens, trims
// leading/trailing hyphens, and truncates to MaxLen. Output may still be
// invalid (empty, reserved, too short) — call Validate to confirm.
func Sanitize(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > MaxLen {
		s = trailingDash.ReplaceAllString(s[:MaxLen], "")
	}
	return s
}

// SanitizeAndValidate sanitizes raw input and runs Validate.
// Returns the sanitized slug on success; empty string + error otherwise.
func SanitizeAndValidate(raw string) (string, error) {
	s := Sanitize(raw)
	if err := Validate(s); err != nil {
		return "", err
	}
	return s, nil
}
