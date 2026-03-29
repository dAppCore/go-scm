// SPDX-Licence-Identifier: EUPL-1.2

package stdio

import (
	"io"
	"syscall"
)

type fdReader struct {
	fd int
}

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

func (w fdWriter) Write(p []byte) (int, error) {
	return syscall.Write(w.fd, p)
}

// Stdin exposes process stdin without importing os.
//
var Stdin io.Reader = fdReader{fd: 0}

// Stdout exposes process stdout without importing os.
//
var Stdout io.Writer = fdWriter{fd: 1}

// Stderr exposes process stderr without importing os.
//
var Stderr io.Writer = fdWriter{fd: 2}
