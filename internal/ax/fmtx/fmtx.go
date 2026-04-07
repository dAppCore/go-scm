// SPDX-License-Identifier: EUPL-1.2

package fmtx

import (
	"fmt"
	"io"

	core "dappco.re/go/core"
	"dappco.re/go/core/scm/internal/ax/stdio"
)

// Sprint mirrors fmt.Sprint using Core primitives.
// Usage: Sprint(...)
func Sprint(args ...any) string {
	return core.Sprint(args...)
}

// Sprintf mirrors fmt.Sprintf using Core primitives.
// Usage: Sprintf(...)
func Sprintf(format string, args ...any) string {
	return core.Sprintf(format, args...)
}

// Fprintf mirrors fmt.Fprintf using Core primitives.
// Usage: Fprintf(...)
func Fprintf(w io.Writer, format string, args ...any) (int, error) {
	return io.WriteString(w, Sprintf(format, args...))
}

// Printf mirrors fmt.Printf.
// Usage: Printf(...)
func Printf(format string, args ...any) (int, error) {
	return Fprintf(stdio.Stdout, format, args...)
}

// Sprintln mirrors fmt.Sprintln — spaces between operands, trailing newline.
// Usage: Sprintln(...)
func Sprintln(args ...any) string {
	return fmt.Sprintln(args...)
}

// Println mirrors fmt.Println — spaces between operands, trailing newline.
// Usage: Println(...)
func Println(args ...any) (int, error) {
	return io.WriteString(stdio.Stdout, Sprintln(args...))
}
