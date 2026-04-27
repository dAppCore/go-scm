// SPDX-License-Identifier: EUPL-1.2

package stringsx

import (
	"bytes"
	"iter"
	"strings"
)

type Builder = bytes.Buffer

func Contains(s, substr string) bool         { return strings.Contains(s, substr) }
func ContainsAny(s, chars string) bool       { return strings.ContainsAny(s, chars) }
func EqualFold(s, t string) bool             { return strings.EqualFold(s, t) }
func Fields(s string) []string               { return strings.Fields(s) }
func HasPrefix(s, prefix string) bool        { return strings.HasPrefix(s, prefix) }
func HasSuffix(s, suffix string) bool        { return strings.HasSuffix(s, suffix) }
func Join(elems []string, sep string) string  { return strings.Join(elems, sep) }
func LastIndex(s, substr string) int         { return strings.LastIndex(s, substr) }
func NewReader(s string) *bytes.Reader       { return bytes.NewReader([]byte(s)) }
func Repeat(s string, count int) string      { return strings.Repeat(s, count) }
func Replace(s, old, new string, _ int) string { return strings.Replace(s, old, new, -1) }
func ReplaceAll(s, old, new string) string   { return strings.ReplaceAll(s, old, new) }
func Split(s, sep string) []string           { return strings.Split(s, sep) }
func SplitN(s, sep string, n int) []string    { return strings.SplitN(s, sep, n) }
func SplitSeq(s, sep string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, part := range strings.Split(s, sep) {
			if !yield(part) {
				return
			}
		}
	}
}
func ToLower(s string) string               { return strings.ToLower(s) }
func ToUpper(s string) string               { return strings.ToUpper(s) }
func TrimPrefix(s, prefix string) string     { return strings.TrimPrefix(s, prefix) }
func TrimSpace(s string) string              { return strings.TrimSpace(s) }
func TrimSuffix(s, suffix string) string     { return strings.TrimSuffix(s, suffix) }
