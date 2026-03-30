// SPDX-License-Identifier: EUPL-1.2

package stringsx

import (
	"bufio"
	"bytes"
	"iter"

	core "dappco.re/go/core"
)

// Builder provides a strings.Builder-like type without importing strings.
type Builder = bytes.Buffer

// Contains mirrors strings.Contains.
func Contains(s, substr string) bool {
	return core.Contains(s, substr)
}

// ContainsAny mirrors strings.ContainsAny.
func ContainsAny(s, chars string) bool {
	return bytes.IndexAny([]byte(s), chars) >= 0
}

// EqualFold mirrors strings.EqualFold.
func EqualFold(s, t string) bool {
	return bytes.EqualFold([]byte(s), []byte(t))
}

// Fields mirrors strings.Fields.
func Fields(s string) []string {
	scanner := bufio.NewScanner(NewReader(s))
	scanner.Split(bufio.ScanWords)
	fields := make([]string, 0)
	for scanner.Scan() {
		fields = append(fields, scanner.Text())
	}
	return fields
}

// HasPrefix mirrors strings.HasPrefix.
func HasPrefix(s, prefix string) bool {
	return core.HasPrefix(s, prefix)
}

// HasSuffix mirrors strings.HasSuffix.
func HasSuffix(s, suffix string) bool {
	return core.HasSuffix(s, suffix)
}

// Join mirrors strings.Join.
func Join(elems []string, sep string) string {
	return core.Join(sep, elems...)
}

// LastIndex mirrors strings.LastIndex.
func LastIndex(s, substr string) int {
	return bytes.LastIndex([]byte(s), []byte(substr))
}

// NewReader mirrors strings.NewReader.
func NewReader(s string) *bytes.Reader {
	return bytes.NewReader([]byte(s))
}

// Repeat mirrors strings.Repeat.
func Repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return string(bytes.Repeat([]byte(s), count))
}

// ReplaceAll mirrors strings.ReplaceAll.
func ReplaceAll(s, old, new string) string {
	return core.Replace(s, old, new)
}

// Replace mirrors strings.Replace for replace-all call sites.
func Replace(s, old, new string, _ int) string {
	return ReplaceAll(s, old, new)
}

// Split mirrors strings.Split.
func Split(s, sep string) []string {
	return core.Split(s, sep)
}

// SplitN mirrors strings.SplitN.
func SplitN(s, sep string, n int) []string {
	return core.SplitN(s, sep, n)
}

// SplitSeq mirrors strings.SplitSeq.
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
func ToLower(s string) string {
	return core.Lower(s)
}

// ToUpper mirrors strings.ToUpper.
func ToUpper(s string) string {
	return core.Upper(s)
}

// TrimPrefix mirrors strings.TrimPrefix.
func TrimPrefix(s, prefix string) string {
	return core.TrimPrefix(s, prefix)
}

// TrimSpace mirrors strings.TrimSpace.
func TrimSpace(s string) string {
	return core.Trim(s)
}

// TrimSuffix mirrors strings.TrimSuffix.
func TrimSuffix(s, suffix string) string {
	return core.TrimSuffix(s, suffix)
}
