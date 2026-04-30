// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: unicode helpers classify whitespace and title-case runes without a banned text import.
	"unicode"

	core "dappco.re/go"
)

func textFields(s string) []string {
	var fields []string
	start := -1
	for i, r := range s {
		if unicode.IsSpace(r) {
			if start >= 0 {
				fields = append(fields, s[start:i])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		fields = append(fields, s[start:])
	}
	return fields
}

func splitTextBySeparators(s string, isSeparator func(rune) bool) []string {
	if isSeparator == nil {
		return []string{s}
	}
	var fields []string
	start := -1
	for i, r := range s {
		if isSeparator(r) {
			if start >= 0 {
				fields = append(fields, s[start:i])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		fields = append(fields, s[start:])
	}
	return fields
}

func equalTextFold(a, b string) bool {
	return core.Lower(a) == core.Lower(b)
}

func replaceFirst(s, old, replacement string) string {
	i := textIndex(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + replacement + s[i+len(old):]
}

func textIndex(s, substr string) int {
	if substr == "" {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func titleText(s string) string {
	runes := []rune(core.Trim(s))
	capNext := true
	for i, r := range runes {
		if r == '-' || r == '_' || r == '/' || unicode.IsSpace(r) {
			capNext = true
			continue
		}
		if capNext {
			runes[i] = unicode.ToTitle(r)
			capNext = false
		}
	}
	return string(runes)
}
