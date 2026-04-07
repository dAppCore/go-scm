// SPDX-License-Identifier: EUPL-1.2

package stringsx

import (
	"bytes"
	"iter"
	"strings"

	core "dappco.re/go/core"
)

// Builder provides a strings.Builder-like type without importing strings.
type Builder = bytes.Buffer

// Contains mirrors strings.Contains.
// Usage: Contains(...)
func Contains(s, substr string) bool {
	return core.Contains(s, substr)
}

// ContainsAny mirrors strings.ContainsAny.
// Usage: ContainsAny(...)
func ContainsAny(s, chars string) bool {
	return bytes.IndexAny([]byte(s), chars) >= 0
}

// EqualFold mirrors strings.EqualFold.
// Usage: EqualFold(...)
func EqualFold(s, t string) bool {
	return bytes.EqualFold([]byte(s), []byte(t))
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
	return bytes.LastIndex([]byte(s), []byte(substr))
}

// NewReader mirrors strings.NewReader.
// Usage: NewReader(...)
func NewReader(s string) *bytes.Reader {
	return bytes.NewReader([]byte(s))
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

// SplitSeq mirrors strings.SplitSeq.
// Usage: SplitSeq(...)
func SplitSeq(s, sep string) iter.Seq[string] {
	parts := Split(s, sep)
	return func(yield func(string) bool) {
		for _, part := range parts {
			if !yield(part) {
				return
			}
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
