package collect

import (
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_Get_Good_ReturnsCopy(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/state.json")

	s.Set("test", &StateEntry{Source: "test", Items: 5})

	// Get returns a copy, so mutating it shouldn't affect internal state.
	got, ok := s.Get("test")
	require.True(t, ok)
	got.Items = 999

	again, ok := s.Get("test")
	require.True(t, ok)
	assert.Equal(t, 5, again.Items, "internal state should not be mutated")
}

func TestState_Save_Good_WritesJSON(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/data/state.json")

	s.Set("src-a", &StateEntry{Source: "src-a", Items: 10, LastID: "abc"})

	err := s.Save()
	require.NoError(t, err)

	// Verify the raw JSON was written.
	content, err := m.Read("/data/state.json")
	require.NoError(t, err)
	assert.Contains(t, content, `"src-a"`)
	assert.Contains(t, content, `"abc"`)
}

func TestState_Load_Good_NullJSON(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/state.json"] = "null"

	s := NewState(m, "/state.json")
	err := s.Load()
	require.NoError(t, err)

	// Null JSON should result in empty entries.
	_, ok := s.Get("anything")
	assert.False(t, ok)
}

func TestState_SaveLoad_Good_WithCursor(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/state.json")

	s.Set("paginated", &StateEntry{
		Source: "paginated",
		Items:  50,
		Cursor: "page_token_abc123",
	})

	err := s.Save()
	require.NoError(t, err)

	s2 := NewState(m, "/state.json")
	err = s2.Load()
	require.NoError(t, err)

	entry, ok := s2.Get("paginated")
	require.True(t, ok)
	assert.Equal(t, "page_token_abc123", entry.Cursor)
}
