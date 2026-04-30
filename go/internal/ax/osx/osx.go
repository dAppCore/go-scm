// SPDX-License-Identifier: EUPL-1.2

package osx

import (
	"io"
	"io/fs"
	"os"
)

func Getenv(key string) string                  { return os.Getenv(key) }
func Getwd() (string, error)                    { return os.Getwd() }
func IsNotExist(err error) bool                 { return os.IsNotExist(err) }
func MkdirAll(path string, _ fs.FileMode) error { return os.MkdirAll(path, 0o755) }
func Open(path string) (fs.File, error)         { return os.Open(path) }
func OpenFile(path string, flag int, _ fs.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(path, flag, 0o600)
}
func ReadDir(path string) ([]fs.DirEntry, error) { return os.ReadDir(path) }
func ReadFile(path string) ([]byte, error)       { return os.ReadFile(path) }
func Stat(path string) (fs.FileInfo, error)      { return os.Stat(path) }
func UserHomeDir() (string, error)               { return os.UserHomeDir() }
func WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}
