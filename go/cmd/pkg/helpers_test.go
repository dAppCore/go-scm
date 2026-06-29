// SPDX-License-Identifier: EUPL-1.2

package pkg

import (
	core "dappco.re/go"
	"dappco.re/go/scm/marketplace"
)

func TestCmdPkg_uniqueCategories_Good(t *core.T) {
	got := uniqueCategories([]string{"provider", "tool"}, []string{"tool", "agent"})
	core.AssertEqual(t, []string{"provider", "tool", "agent"}, got)
}

func TestCmdPkg_uniqueCategories_Bad(t *core.T) {
	got := uniqueCategories(nil, nil)
	core.AssertEmpty(t, got)
}

func TestCmdPkg_uniqueCategories_Ugly(t *core.T) {
	got := uniqueCategories([]string{" ", "provider"}, []string{"  "})
	core.AssertEqual(t, []string{"provider"}, got)
}

func TestCmdPkg_splitList_Good(t *core.T) {
	got := splitList("a, b ,c")
	core.AssertEqual(t, []string{"a", "b", "c"}, got)
}

func TestCmdPkg_splitList_Bad(t *core.T) {
	core.AssertNil(t, splitList("   "))
}

func TestCmdPkg_splitList_Ugly(t *core.T) {
	got := splitList(",,a,,")
	core.AssertEqual(t, []string{"a"}, got)
}

func TestCmdPkg_trimRightSlash_Good(t *core.T) {
	core.AssertEqual(t, "https://forge.example", trimRightSlash("https://forge.example/"))
}

func TestCmdPkg_trimRightSlash_Bad(t *core.T) {
	core.AssertEqual(t, "https://forge.example", trimRightSlash("https://forge.example"))
}

func TestCmdPkg_trimRightSlash_Ugly(t *core.T) {
	core.AssertEqual(t, "https://forge.example", trimRightSlash("https://forge.example///"))
}

func TestCmdPkg_sortIndex_Good(t *core.T) {
	idx := &marketplace.Index{
		Modules:    []marketplace.Module{{Code: "zeta"}, {Code: "alpha"}},
		Categories: []string{"tool", "agent"},
	}
	sortIndex(idx)
	core.AssertEqual(t, "alpha", idx.Modules[0].Code)
	core.AssertEqual(t, "agent", idx.Categories[0])
}

func TestCmdPkg_sortIndex_Bad(t *core.T) {
	sortIndex(nil)
}

func TestCmdPkg_sortIndex_Ugly(t *core.T) {
	idx := &marketplace.Index{}
	sortIndex(idx)
	core.AssertEmpty(t, idx.Modules)
}

func TestCmdPkg_applyRepoDefaults_Good(t *core.T) {
	idx := &marketplace.Index{Modules: []marketplace.Module{{Code: "demo"}}}
	applyRepoDefaults(idx, "https://forge.example/", "modules")
	core.AssertEqual(t, "https://forge.example/modules/demo", idx.Modules[0].Repo)
}

func TestCmdPkg_applyRepoDefaults_Bad(t *core.T) {
	idx := &marketplace.Index{Modules: []marketplace.Module{{Code: "demo"}}}
	applyRepoDefaults(idx, "", "modules")
	core.AssertEqual(t, "", idx.Modules[0].Repo)
}

func TestCmdPkg_applyRepoDefaults_Ugly(t *core.T) {
	// Empty org falls back to "core".
	idx := &marketplace.Index{Modules: []marketplace.Module{{Code: "demo"}}}
	applyRepoDefaults(idx, "https://forge.example", "")
	core.AssertEqual(t, "https://forge.example/core/demo", idx.Modules[0].Repo)
}

func TestCmdPkg_failed_Good(t *core.T) {
	result := failed(core.E("cmd.pkg", "boom", nil))
	core.AssertFalse(t, result.OK)
}

func TestCmdPkg_failed_Bad(t *core.T) {
	result := failed(nil)
	core.AssertFalse(t, result.OK)
}

func TestCmdPkg_resultError_Good(t *core.T) {
	err := resultError("cmd.pkg", "wrap", core.Fail(core.E("inner", "cause", nil)))
	core.AssertContains(t, err.Error(), "wrap")
}

func TestCmdPkg_resultError_Bad(t *core.T) {
	err := resultError("cmd.pkg", "no cause", core.Ok(nil))
	core.AssertContains(t, err.Error(), "no cause")
}
