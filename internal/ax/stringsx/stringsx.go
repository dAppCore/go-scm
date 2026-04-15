// SPDX-License-Identifier: EUPL-1.2

package stringsx

import (
	"iter"
	"strings"

	core "dappco.re/go/core"
)

// Builder is an alias for strings.Builder for use without importing strings directly.
type Builder = strings.Builder

// Contains mirrors strings.Contains.
// Usage: Contains(...)
func Contains(s, substr string) bool {
	return core.Contains(s, substr)
}

// ContainsAny mirrors strings.ContainsAny.
// Usage: ContainsAny(...)
func ContainsAny(s, chars string) bool {
	return strings.ContainsAny(s, chars)
}

// EqualFold mirrors strings.EqualFold.
// Usage: EqualFold(...)
func EqualFold(s, t string) bool {
	return strings.EqualFold(s, t)
}

// Fields mirrors strings.Fields.
// Usage: Fields(...)
func Fields(s string) []string {
	return strings.Fields(s)
}

// HasPrefix mirrors strings.HasPrefix.
// Usage: HasPrefix(...)
func HasPrefix(s, prefix string) bool {
	return core.HasPrefix(s, prefix)
}

// HasSuffix mirrors strings.HasSuffix.
// Usage: HasSuffix(...)
func HasSuffix(s, suffix string) bool {
	return core.HasSuffix(s, suffix)
}

// Join mirrors strings.Join.
// Usage: Join(...)
func Join(elems []string, sep string) string {
	return core.Join(sep, elems...)
}

// LastIndex mirrors strings.LastIndex.
// Usage: LastIndex(...)
func LastIndex(s, substr string) int {
	return strings.LastIndex(s, substr)
}

// NewReader mirrors strings.NewReader.
// Usage: NewReader(...)
func NewReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

// Repeat mirrors strings.Repeat.
// Usage: Repeat(...)
func Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

// ReplaceAll mirrors strings.ReplaceAll.
// Usage: ReplaceAll(...)
func ReplaceAll(s, old, new string) string {
	return core.Replace(s, old, new)
}

// Replace mirrors strings.Replace.
// Usage: Replace(...)
func Replace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

// Split mirrors strings.Split.
// Usage: Split(...)
func Split(s, sep string) []string {
	return core.Split(s, sep)
}

// SplitN mirrors strings.SplitN.
// Usage: SplitN(...)
func SplitN(s, sep string, n int) []string {
	return core.SplitN(s, sep, n)
}

// SplitSeq mirrors strings.SplitSeq, lazily yielding substrings without
// pre-allocating the full slice so early iteration termination is cheap.
// Usage: SplitSeq(...)
func SplitSeq(s, sep string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if sep == "" {
			for _, r := range s {
				if !yield(string(r)) {
					return
				}
			}
			return
		}
		for {
			idx := strings.Index(s, sep)
			if idx < 0 {
				yield(s)
				return
			}
			if !yield(s[:idx]) {
				return
			}
			s = s[idx+len(sep):]
		}
	}
}

// ToLower mirrors strings.ToLower.
// Usage: ToLower(...)
func ToLower(s string) string {
	return core.Lower(s)
}

// ToUpper mirrors strings.ToUpper.
// Usage: ToUpper(...)
func ToUpper(s string) string {
	return core.Upper(s)
}

// TrimPrefix mirrors strings.TrimPrefix.
// Usage: TrimPrefix(...)
func TrimPrefix(s, prefix string) string {
	return core.TrimPrefix(s, prefix)
}

// TrimSpace mirrors strings.TrimSpace.
// Usage: TrimSpace(...)
func TrimSpace(s string) string {
	return core.Trim(s)
}

// TrimSuffix mirrors strings.TrimSuffix.
// Usage: TrimSuffix(...)
func TrimSuffix(s, suffix string) string {
	return core.TrimSuffix(s, suffix)
}
