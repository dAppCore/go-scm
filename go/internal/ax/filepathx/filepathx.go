// SPDX-License-Identifier: EUPL-1.2

package filepathx

import "path/filepath"

func Abs(p string) (string, error) { return filepath.Abs(p) }
func Base(p string) string         { return filepath.Base(p) }
func Clean(p string) string        { return filepath.Clean(p) }
func Dir(p string) string          { return filepath.Dir(p) }
func Ext(p string) string          { return filepath.Ext(p) }
func Join(elem ...string) string   { return filepath.Join(elem...) }
