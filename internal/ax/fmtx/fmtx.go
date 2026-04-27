// SPDX-License-Identifier: EUPL-1.2

package fmtx

import (
	"fmt"
	"io"
)

func Fprintf(w io.Writer, format string, args ...any) (int, error) { return fmt.Fprintf(w, format, args...) }
func Printf(format string, args ...any) (int, error)               { return fmt.Printf(format, args...) }
func Println(args ...any) (int, error)                             { return fmt.Println(args...) }
func Sprint(args ...any) string                                    { return fmt.Sprint(args...) }
func Sprintf(format string, args ...any) string                    { return fmt.Sprintf(format, args...) }
