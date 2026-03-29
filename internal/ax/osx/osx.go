// SPDX-Licence-Identifier: EUPL-1.2

package osx

import (
	"io"
	"io/fs"
	"os/user"
	"syscall"

	core "dappco.re/go/core"
	coreio "dappco.re/go/core/io"
	"dappco.re/go/core/scm/internal/ax/stdio"
)

const (
	//
	O_APPEND = syscall.O_APPEND
	//
	O_CREATE = syscall.O_CREAT
	//
	O_WRONLY = syscall.O_WRONLY
)

// Stdin exposes process stdin without importing os.
//
var Stdin = stdio.Stdin

// Stdout exposes process stdout without importing os.
//
var Stdout = stdio.Stdout

// Stderr exposes process stderr without importing os.
//
var Stderr = stdio.Stderr

// Getenv mirrors os.Getenv.
//
func Getenv(key string) string {
	value, _ := syscall.Getenv(key)
	return value
}

// Getwd mirrors os.Getwd.
//
func Getwd() (string, error) {
	return syscall.Getwd()
}

// IsNotExist mirrors os.IsNotExist.
//
func IsNotExist(err error) bool {
	return core.Is(err, fs.ErrNotExist)
}

// MkdirAll mirrors os.MkdirAll.
//
func MkdirAll(path string, _ fs.FileMode) error {
	return coreio.Local.EnsureDir(path)
}

// Open mirrors os.Open.
//
func Open(path string) (fs.File, error) {
	return coreio.Local.Open(path)
}

// OpenFile mirrors the append/create/write mode used in this repo.
//
func OpenFile(path string, flag int, _ fs.FileMode) (io.WriteCloser, error) {
	if flag&O_APPEND != 0 {
		return coreio.Local.Append(path)
	}
	return coreio.Local.Create(path)
}

// ReadDir mirrors os.ReadDir.
//
func ReadDir(path string) ([]fs.DirEntry, error) {
	return coreio.Local.List(path)
}

// ReadFile mirrors os.ReadFile.
//
func ReadFile(path string) ([]byte, error) {
	content, err := coreio.Local.Read(path)
	return []byte(content), err
}

// Stat mirrors os.Stat.
//
func Stat(path string) (fs.FileInfo, error) {
	return coreio.Local.Stat(path)
}

// UserHomeDir mirrors os.UserHomeDir.
//
func UserHomeDir() (string, error) {
	if home := Getenv("HOME"); home != "" {
		return home, nil
	}
	current, err := user.Current()
	if err != nil {
		return "", err
	}
	return current.HomeDir, nil
}

// WriteFile mirrors os.WriteFile.
//
func WriteFile(path string, data []byte, perm fs.FileMode) error {
	return coreio.Local.WriteMode(path, string(data), perm)
}
