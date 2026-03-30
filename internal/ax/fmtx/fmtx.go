// SPDX-License-Identifier: EUPL-1.2

package fmtx

import (
	"io"

	core "dappco.re/go/core"
	"dappco.re/go/core/scm/internal/ax/stdio"
)

// Sprint mirrors fmt.Sprint using Core primitives.
func Sprint(args ...any) string {
	return core.Sprint(args...)
}

// Sprintf mirrors fmt.Sprintf using Core primitives.
func Sprintf(format string, args ...any) string {
	return core.Sprintf(format, args...)
}

// Fprintf mirrors fmt.Fprintf using Core primitives.
func Fprintf(w io.Writer, format string, args ...any) (int, error) {
	return io.WriteString(w, Sprintf(format, args...))
}

// Printf mirrors fmt.Printf.
func Printf(format string, args ...any) (int, error) {
	return Fprintf(stdio.Stdout, format, args...)
}

// Println mirrors fmt.Println.
func Println(args ...any) (int, error) {
	return io.WriteString(stdio.Stdout, Sprint(args...)+"\n")
}
