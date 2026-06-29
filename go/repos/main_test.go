// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"testing"

	core "dappco.re/go"
	process "dappco.re/go/process"
)

// TestMain initialises the global process service before the package's
// tests run. The registry sync paths shell out via process.RunWithOptions,
// which v0.10.3 requires be initialised through process.Init first.
//
// Example:
//
//	func TestMain(m *testing.M) { ... process.Init(core.New(...)) ... }
func TestMain(m *testing.M) {
	c := core.New(core.WithOption("name", "repos-test"))
	if r := process.Init(c); !r.OK {
		core.Error("repos test setup failed", "err", r.Error())
		core.Exit(1)
		return
	}
	core.Exit(m.Run())
}
