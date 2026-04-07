// SPDX-License-Identifier: EUPL-1.2

package syncutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepoName_Good(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "RepoOnly", input: "core", want: "core"},
		{name: "OwnerRepo", input: "host-uk/core", want: "core"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepoName(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseRepoName_Bad(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "PathTraversal", input: "../escape"},
		{name: "PathTraversalEncoded", input: "host-uk%2F..%2Fescape"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRepoName(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "syncutil.ParseRepoName")
		})
	}
}
