// SPDX-License-Identifier: EUPL-1.2

package stdio

import (
	"io"
	"syscall"
)

type fdReader struct {
	fd int
}

// Read implements io.Reader for stdin without importing os.
// Usage: Read(...)
func (r fdReader) Read(p []byte) (int, error) {
	n, err := syscall.Read(r.fd, p)
	if n == 0 && err == nil {
		return 0, io.EOF
	}
	return n, err
}

type fdWriter struct {
	fd int
}

// Write implements io.Writer for stdout and stderr without importing os.
// Usage: Write(...)
func (w fdWriter) Write(p []byte) (int, error) {
	n, err := syscall.Write(w.fd, p)
	if n < len(p) && err == nil {
		return n, io.ErrShortWrite
	}
	return n, err
}

// Stdin exposes process stdin without importing os.
var Stdin io.Reader = fdReader{fd: 0}

// Stdout exposes process stdout without importing os.
var Stdout io.Writer = fdWriter{fd: 1}

// Stderr exposes process stderr without importing os.
var Stderr io.Writer = fdWriter{fd: 2}
