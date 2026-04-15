// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"testing"

	coreio "dappco.re/go/core/io"
)

func TestNewStateWithEmptyPathDoesNotPersist(t *testing.T) {
	medium := coreio.NewMockMedium()
	state := NewState(medium, "")

	if err := state.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if len(medium.Files) != 0 {
		t.Fatalf("expected empty path state to skip persistence, got %#v", medium.Files)
	}
}
