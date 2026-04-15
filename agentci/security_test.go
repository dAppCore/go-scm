// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"context"
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
		{"parent traversal", "../secret"},
		{"absolute path", "/var/tmp/report.txt"},
		{"nested path", "nested/path/file"},
		{"spaces", "has space"},
		{"special chars", "file@name"},
		{"backtick", "file`name"},
		{"semicolon", "file;name"},
		{"pipe", "file|name"},
		{"ampersand", "file&name"},
		{"dollar", "file$name"},
		{"backslash", `path\to\file.txt`},
		{"current dir", "."},
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

func TestValidateRemoteDir_Good(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/", "/"},
		{"~", "~"},
		{"queue", "queue"},
		{"queue/subdir", "queue/subdir"},
		{"/var/tmp/queue", "/var/tmp/queue"},
		{"~/ai-work/queue", "~/ai-work/queue"},
		{"queue//nested", "queue/nested"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ValidateRemoteDir(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateRemoteDir_Bad(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"parent traversal", "../queue"},
		{"nested traversal", "queue/../done"},
		{"absolute traversal", "/var/../tmp"},
		{"home traversal", "~/ai-work/../../queue"},
		{"dot segment", "queue/./nested"},
		{"backslash", `queue\done`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateRemoteDir(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestJoinRemotePath_Good_BaseOnly_Good(t *testing.T) {
	got, err := JoinRemotePath("~/ai-work/queue")
	require.NoError(t, err)
	assert.Equal(t, "~/ai-work/queue", got)
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

func TestSecureSSHCommandContext_Good(t *testing.T) {
	cmd := SecureSSHCommandContext(context.Background(), "host.example.com", "ls -la")
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
