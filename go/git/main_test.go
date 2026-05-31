// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"testing"

	core "dappco.re/go"
	process "dappco.re/go/process"
)

// TestMain initialises the global process service before the package's
// tests run. The git helpers call process.RunWithOptions, which requires
// the default service to be initialised via process.Init.
//
// Example:
//
//	func TestMain(m *testing.M) { ... process.Init(core.New(...)) ... }
func TestMain(m *testing.M) {
	c := core.New(core.WithOption("name", "git-test"))
	if r := process.Init(c); !r.OK {
		core.Error("git test setup failed", "err", r.Error())
		core.Exit(1)
		return
	}
	core.Exit(m.Run())
}
