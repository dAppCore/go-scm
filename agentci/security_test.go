// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizePath_Good(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"with-dash", "with-dash"},
		{"with_underscore", "with_underscore"},
		{"with.dot", "with.dot"},
		{"CamelCase", "CamelCase"},
		{"123", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := SanitizePath(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizePath_Bad(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"spaces", "has space"},
		{"special chars", "file@name"},
		{"backtick", "file`name"},
		{"semicolon", "file;name"},
		{"pipe", "file|name"},
		{"ampersand", "file&name"},
		{"dollar", "file$name"},
		{"slash", "path/to/file.txt"},
		{"backslash", `path\to\file.txt`},
		{"parent traversal base", ".."},
		{"root", "/"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SanitizePath(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestEscapeShellArg_Good(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "'simple'"},
		{"with spaces", "'with spaces'"},
		{"it's", "'it'\\''s'"},
		{"", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, EscapeShellArg(tt.input))
		})
	}
}

func TestSecureSSHCommand_Good(t *testing.T) {
	cmd := SecureSSHCommand("host.example.com", "ls -la")
	args := cmd.Args

	assert.Equal(t, "ssh", args[0])
	assert.Contains(t, args, "-o")
	assert.Contains(t, args, "StrictHostKeyChecking=yes")
	assert.Contains(t, args, "BatchMode=yes")
	assert.Contains(t, args, "ConnectTimeout=10")
	assert.Equal(t, "host.example.com", args[len(args)-2])
	assert.Equal(t, "ls -la", args[len(args)-1])
}

func TestMaskToken_Good(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"long token", "abcdefghijklmnop", "abcd****mnop"},
		{"exactly 8", "12345678", "1234****5678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskToken(tt.input))
		})
	}
}

func TestMaskToken_Bad(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"short", "abc"},
		{"empty", ""},
		{"seven chars", "1234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, "*****", MaskToken(tt.input))
		})
	}
}
