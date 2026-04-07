// SPDX-License-Identifier: EUPL-1.2

package filepathx

import (
	"fmt"
	"path"
	"path/filepath"
	"syscall"
)

// Separator mirrors filepath.Separator for Unix-style Core paths.
const Separator = '/'

// Abs mirrors filepath.Abs for the paths used in this repo.
// Usage: Abs(...)
func Abs(p string) (string, error) {
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	cwd, err := syscall.Getwd()
	if err != nil {
		return "", fmt.Errorf("filepathx.Abs: %w", err)
	}
	return filepath.Clean(filepath.Join(cwd, p)), nil
}

// Base mirrors filepath.Base.
// Usage: Base(...)
func Base(p string) string {
	return path.Base(p)
}

// Clean mirrors filepath.Clean.
// Usage: Clean(...)
func Clean(p string) string {
	return path.Clean(p)
}

// Dir mirrors filepath.Dir.
// Usage: Dir(...)
func Dir(p string) string {
	return path.Dir(p)
}

// Ext mirrors filepath.Ext.
// Usage: Ext(...)
func Ext(p string) string {
	return path.Ext(p)
}

// Join mirrors filepath.Join.
// Usage: Join(...)
func Join(elem ...string) string {
	return path.Join(elem...)
}
