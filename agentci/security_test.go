// SPDX-License-Identifier: EUPL-1.2

package agentci

import (
	"context"
	"slices"
	"testing"
)

func checkNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func checkError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
}

func checkEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func checkContains[T comparable](t *testing.T, items []T, want T) {
	t.Helper()
	if !slices.Contains(items, want) {
		t.Fatalf("expected %v to contain %v", items, want)
	}
}

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
			checkNoError(t, err)
			checkEqual(t, tt.want, got)
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
			checkError(t, err)
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
			checkEqual(t, tt.want, EscapeShellArg(tt.input))
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
		{"~/", "~"},
		{"queue", "queue"},
		{"queue/subdir", "queue/subdir"},
		{"/var/tmp/queue", "/var/tmp/queue"},
		{"~/ai-work/queue", "~/ai-work/queue"},
		{"queue//nested", "queue/nested"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ValidateRemoteDir(tt.input)
			checkNoError(t, err)
			checkEqual(t, tt.want, got)
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
			checkError(t, err)
		})
	}
}

func TestJoinRemotePath_Good_BaseOnly_Good(t *testing.T) {
	got, err := JoinRemotePath("~/ai-work/queue")
	checkNoError(t, err)
	checkEqual(t, "~/ai-work/queue", got)
}

func TestResolvePathWithinRoot_Good_RootDirectory_Good(t *testing.T) {
	safe, resolved, err := ResolvePathWithinRoot("/", "tmp")
	checkNoError(t, err)
	checkEqual(t, "tmp", safe)
	checkEqual(t, "/tmp", resolved)
}

func TestResolvePathWithinRoot_Good_Subdirectory_Good(t *testing.T) {
	safe, resolved, err := ResolvePathWithinRoot("/var/lib", "core")
	checkNoError(t, err)
	checkEqual(t, "core", safe)
	checkEqual(t, "/var/lib/core", resolved)
}

func TestResolvePathWithinRoot_Bad_EscapesRoot(t *testing.T) {
	_, _, err := ResolvePathWithinRoot("/var/lib", "../core")
	checkError(t, err)
}

func TestResolvePathWithinRoot_Bad_EmptyRoot(t *testing.T) {
	_, _, err := ResolvePathWithinRoot("", "core")
	checkError(t, err)
}

func TestSecureSSHCommand_Good(t *testing.T) {
	cmd := SecureSSHCommand("host.example.com", "ls -la")
	args := cmd.Args

	checkEqual(t, "ssh", args[0])
	checkContains(t, args, "-o")
	checkContains(t, args, "StrictHostKeyChecking=yes")
	checkContains(t, args, "BatchMode=yes")
	checkContains(t, args, "ConnectTimeout=10")
	checkEqual(t, "host.example.com", args[len(args)-2])
	checkEqual(t, "ls -la", args[len(args)-1])
}

func TestSecureSSHCommandContext_Good(t *testing.T) {
	cmd := SecureSSHCommandContext(context.Background(), "host.example.com", "ls -la")
	args := cmd.Args

	checkEqual(t, "ssh", args[0])
	checkContains(t, args, "-o")
	checkContains(t, args, "StrictHostKeyChecking=yes")
	checkContains(t, args, "BatchMode=yes")
	checkContains(t, args, "ConnectTimeout=10")
	checkEqual(t, "host.example.com", args[len(args)-2])
	checkEqual(t, "ls -la", args[len(args)-1])
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
			checkEqual(t, tt.want, MaskToken(tt.input))
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
			checkEqual(t, "*****", MaskToken(tt.input))
		})
	}
}
