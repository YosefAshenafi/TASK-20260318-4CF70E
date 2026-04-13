package pii

import "strings"

// PartialMaskPhone shows only the last few digits (list/detail masking when full PII not granted).
func PartialMaskPhone(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r := []rune(s)
	if len(r) <= 4 {
		return "••••"
	}
	return "••••••" + string(r[len(r)-4:])
}

// PartialMaskID shows only the last few characters of an ID number.
func PartialMaskID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r := []rune(s)
	if len(r) <= 4 {
		return "••••"
	}
	return "••••••" + string(r[len(r)-4:])
}

// PartialMaskEmail masks the local part (design §11.2 list-style desensitization).
func PartialMaskEmail(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	parts := strings.SplitN(s, "@", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "••••"
	}
	local := []rune(parts[0])
	if len(local) == 0 {
		return "••••@" + parts[1]
	}
	return string(local[0:1]) + "•••@" + parts[1]
}
