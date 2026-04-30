// SPDX-License-Identifier: EUPL-1.2

package filepathx

import core "dappco.re/go"

func TestFilepathx_Abs_Good(t *core.T) {
	got, err := Abs(".")
	core.AssertNoError(t, err)
	core.AssertTrue(t, got != "")
}

func TestFilepathx_Abs_Bad(t *core.T) {
	got, err := Abs("")
	core.AssertNoError(t, err)
	core.AssertTrue(t, got != "")
}

func TestFilepathx_Abs_Ugly(t *core.T) {
	got, err := Abs(t.TempDir())
	core.AssertNoError(t, err)
	core.AssertTrue(t, Base(got) != "")
}

func TestFilepathx_Base_Good(t *core.T) {
	got := Base("/tmp/agent.yaml")
	core.AssertEqual(
		t, "agent.yaml", got,
	)
}

func TestFilepathx_Base_Bad(t *core.T) {
	got := Base("")
	core.AssertEqual(
		t, ".", got,
	)
}

func TestFilepathx_Base_Ugly(t *core.T) {
	got := Base("/")
	core.AssertEqual(
		t, "/", got,
	)
}

func TestFilepathx_Clean_Good(t *core.T) {
	got := Clean("repo/../repo/.core")
	core.AssertEqual(
		t, "repo/.core", got,
	)
}

func TestFilepathx_Clean_Bad(t *core.T) {
	got := Clean("")
	core.AssertEqual(
		t, ".", got,
	)
}

func TestFilepathx_Clean_Ugly(t *core.T) {
	got := Clean("../../repo")
	core.AssertEqual(
		t, "../../repo", got,
	)
}

func TestFilepathx_Dir_Good(t *core.T) {
	got := Dir("/tmp/repo/core.json")
	core.AssertEqual(
		t, "/tmp/repo", got,
	)
}

func TestFilepathx_Dir_Bad(t *core.T) {
	got := Dir("core.json")
	core.AssertEqual(
		t, ".", got,
	)
}

func TestFilepathx_Dir_Ugly(t *core.T) {
	got := Dir("/")
	core.AssertEqual(
		t, "/", got,
	)
}

func TestFilepathx_Ext_Good(t *core.T) {
	got := Ext("manifest.yaml")
	core.AssertEqual(
		t, ".yaml", got,
	)
}

func TestFilepathx_Ext_Bad(t *core.T) {
	got := Ext("manifest")
	core.AssertEqual(
		t, "", got,
	)
}

func TestFilepathx_Ext_Ugly(t *core.T) {
	got := Ext(".gitignore")
	core.AssertEqual(
		t, ".gitignore", got,
	)
}

func TestFilepathx_Join_Good(t *core.T) {
	got := Join("repo", ".core", "manifest.yaml")
	core.AssertEqual(
		t, "repo/.core/manifest.yaml", got,
	)
}

func TestFilepathx_Join_Bad(t *core.T) {
	got := Join()
	core.AssertEqual(
		t, "", got,
	)
}

func TestFilepathx_Join_Ugly(t *core.T) {
	got := Join("repo", "..", "repo", "core.json")
	core.AssertEqual(
		t, "repo/core.json", got,
	)
}
