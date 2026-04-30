// SPDX-License-Identifier: EUPL-1.2

package fmtx

import core "dappco.re/go"

const (
	sonarFmtxTestAgentCodex = "agent=codex"
	sonarFmtxTestAgentS     = "agent=%s"
)

func TestFmtx_Fprintf_Good(t *core.T) {
	builder := core.NewBuilder()
	n, err := Fprintf(builder, sonarFmtxTestAgentS, "codex")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 11, n)
	core.AssertEqual(t, sonarFmtxTestAgentCodex, builder.String())
}

func TestFmtx_Fprintf_Bad(t *core.T) {
	builder := core.NewBuilder()
	format := "%d"
	n, err := Fprintf(builder, format, "codex")
	core.AssertNoError(t, err)
	core.AssertTrue(t, n > 0)
	core.AssertContains(t, builder.String(), "%!d")
}

func TestFmtx_Fprintf_Ugly(t *core.T) {
	builder := core.NewBuilder()
	n, err := Fprintf(builder, "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 0, n)
}

func TestFmtx_Printf_Good(t *core.T) {
	n, err := Printf(sonarFmtxTestAgentS, "codex")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 11, n)
}

func TestFmtx_Printf_Bad(t *core.T) {
	format := "%d"
	n, err := Printf(format, "codex")
	core.AssertNoError(t, err)
	core.AssertTrue(t, n > 0)
}

func TestFmtx_Printf_Ugly(t *core.T) {
	n, err := Printf("")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 0, n)
}

func TestFmtx_Println_Good(t *core.T) {
	n, err := Println("agent", "codex")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 12, n)
}

func TestFmtx_Println_Bad(t *core.T) {
	n, err := Println()
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, n)
}

func TestFmtx_Println_Ugly(t *core.T) {
	n, err := Println("")
	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, n)
}

func TestFmtx_Sprint_Good(t *core.T) {
	got := Sprint("agent", "=", "codex")
	core.AssertEqual(
		t, sonarFmtxTestAgentCodex, got,
	)
}

func TestFmtx_Sprint_Bad(t *core.T) {
	got := Sprint()
	core.AssertEqual(
		t, "", got,
	)
}

func TestFmtx_Sprint_Ugly(t *core.T) {
	got := Sprint(nil)
	core.AssertEqual(
		t, "<nil>", got,
	)
}

func TestFmtx_Sprintf_Good(t *core.T) {
	got := Sprintf(sonarFmtxTestAgentS, "codex")
	core.AssertEqual(
		t, sonarFmtxTestAgentCodex, got,
	)
}

func TestFmtx_Sprintf_Bad(t *core.T) {
	format := "%d"
	got := Sprintf(format, "codex")
	core.AssertContains(t, got, "%!d")
}

func TestFmtx_Sprintf_Ugly(t *core.T) {
	got := Sprintf("")
	core.AssertEqual(
		t, "", got,
	)
}
