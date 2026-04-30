// SPDX-License-Identifier: EUPL-1.2

package stringsx

import core "dappco.re/go"

const (
	sonarStringsxTestAgentDispatch  = "agent.dispatch"
	sonarStringsxTestAgentDispatch2 = "agent/dispatch"
	sonarStringsxTestManifestYaml   = "manifest.yaml"
)

func TestStringsx_Contains_Good(t *core.T) {
	got := Contains(sonarStringsxTestAgentDispatch, "dispatch")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_Contains_Bad(t *core.T) {
	got := Contains(sonarStringsxTestAgentDispatch, "missing")
	core.AssertFalse(
		t, got,
	)
}

func TestStringsx_Contains_Ugly(t *core.T) {
	got := Contains("", "")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_ContainsAny_Good(t *core.T) {
	got := ContainsAny(sonarStringsxTestAgentDispatch, ".:")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_ContainsAny_Bad(t *core.T) {
	got := ContainsAny("agent", "0123")
	core.AssertFalse(
		t, got,
	)
}

func TestStringsx_ContainsAny_Ugly(t *core.T) {
	got := ContainsAny("", "abc")
	core.AssertFalse(
		t, got,
	)
}

func TestStringsx_EqualFold_Good(t *core.T) {
	got := EqualFold("ForgeJo", "forgejo")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_EqualFold_Bad(t *core.T) {
	got := EqualFold("forge", "gitea")
	core.AssertFalse(
		t, got,
	)
}

func TestStringsx_EqualFold_Ugly(t *core.T) {
	got := EqualFold("", "")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_Fields_Good(t *core.T) {
	got := Fields("agent dispatch ready")
	core.AssertEqual(
		t, []string{"agent", "dispatch", "ready"}, got,
	)
}

func TestStringsx_Fields_Bad(t *core.T) {
	got := Fields("   ")
	core.AssertEmpty(
		t, got,
	)
}

func TestStringsx_Fields_Ugly(t *core.T) {
	got := Fields("\nagent\tready\r\n")
	core.AssertEqual(
		t, []string{"agent", "ready"}, got,
	)
}

func TestStringsx_HasPrefix_Good(t *core.T) {
	got := HasPrefix(sonarStringsxTestAgentDispatch, "agent")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_HasPrefix_Bad(t *core.T) {
	got := HasPrefix(sonarStringsxTestAgentDispatch, "task")
	core.AssertFalse(
		t, got,
	)
}

func TestStringsx_HasPrefix_Ugly(t *core.T) {
	got := HasPrefix(sonarStringsxTestAgentDispatch, "")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_HasSuffix_Good(t *core.T) {
	got := HasSuffix(sonarStringsxTestManifestYaml, ".yaml")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_HasSuffix_Bad(t *core.T) {
	got := HasSuffix(sonarStringsxTestManifestYaml, ".json")
	core.AssertFalse(
		t, got,
	)
}

func TestStringsx_HasSuffix_Ugly(t *core.T) {
	got := HasSuffix(sonarStringsxTestManifestYaml, "")
	core.AssertTrue(
		t, got,
	)
}

func TestStringsx_Join_Good(t *core.T) {
	got := Join([]string{"agent", "dispatch"}, ".")
	core.AssertEqual(
		t, sonarStringsxTestAgentDispatch, got,
	)
}

func TestStringsx_Join_Bad(t *core.T) {
	got := Join(nil, ".")
	core.AssertEqual(
		t, "", got,
	)
}

func TestStringsx_Join_Ugly(t *core.T) {
	got := Join([]string{"agent", "", "ready"}, "/")
	core.AssertEqual(
		t, "agent//ready", got,
	)
}

func TestStringsx_LastIndex_Good(t *core.T) {
	got := LastIndex("agent.dispatch.ready", ".")
	core.AssertEqual(
		t, 14, got,
	)
}

func TestStringsx_LastIndex_Bad(t *core.T) {
	got := LastIndex(sonarStringsxTestAgentDispatch, "/")
	core.AssertEqual(
		t, -1, got,
	)
}

func TestStringsx_LastIndex_Ugly(t *core.T) {
	got := LastIndex(sonarStringsxTestAgentDispatch, "")
	core.AssertEqual(
		t, len(sonarStringsxTestAgentDispatch), got,
	)
}

func TestStringsx_NewReader_Good(t *core.T) {
	reader := NewReader("agent")
	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	core.AssertNoError(t, err)
	core.AssertEqual(t, 5, n)
	core.AssertEqual(t, "agent", string(buf))
}

func TestStringsx_NewReader_Bad(t *core.T) {
	reader := NewReader("")
	buf := make([]byte, 1)
	n, err := reader.Read(buf)
	core.AssertEqual(t, 0, n)
	core.AssertErrorIs(t, err, core.EOF)
}

func TestStringsx_NewReader_Ugly(t *core.T) {
	reader := NewReader("agent dispatch")
	offset, err := reader.Seek(6, 0)
	core.RequireNoError(t, err)
	core.AssertEqual(t, int64(6), offset)
}

func TestStringsx_Repeat_Good(t *core.T) {
	got := Repeat("go", 3)
	core.AssertEqual(
		t, "gogogo", got,
	)
}

func TestStringsx_Repeat_Bad(t *core.T) {
	got := Repeat("go", 0)
	core.AssertEqual(
		t, "", got,
	)
}

func TestStringsx_Repeat_Ugly(t *core.T) {
	core.AssertPanics(t, func() {
		_ = Repeat("go", -1)
	})
}

func TestStringsx_Replace_Good(t *core.T) {
	got := Replace(sonarStringsxTestAgentDispatch2, "/", ".", 1)
	core.AssertEqual(
		t, sonarStringsxTestAgentDispatch, got,
	)
}

func TestStringsx_Replace_Bad(t *core.T) {
	got := Replace(sonarStringsxTestAgentDispatch2, ".", "/", 1)
	core.AssertEqual(
		t, sonarStringsxTestAgentDispatch2, got,
	)
}

func TestStringsx_Replace_Ugly(t *core.T) {
	got := Replace("agent", "", ".", 1)
	core.AssertEqual(
		t, ".a.g.e.n.t.", got,
	)
}

func TestStringsx_ReplaceAll_Good(t *core.T) {
	got := ReplaceAll("agent/dispatch/run", "/", ".")
	core.AssertEqual(
		t, "agent.dispatch.run", got,
	)
}

func TestStringsx_ReplaceAll_Bad(t *core.T) {
	got := ReplaceAll("agent", "missing", "x")
	core.AssertEqual(
		t, "agent", got,
	)
}

func TestStringsx_ReplaceAll_Ugly(t *core.T) {
	got := ReplaceAll("", "", ".")
	core.AssertEqual(
		t, ".", got,
	)
}

func TestStringsx_Split_Good(t *core.T) {
	got := Split(sonarStringsxTestAgentDispatch2, "/")
	core.AssertEqual(
		t, []string{"agent", "dispatch"}, got,
	)
}

func TestStringsx_Split_Bad(t *core.T) {
	got := Split(sonarStringsxTestAgentDispatch2, ".")
	core.AssertEqual(
		t, []string{sonarStringsxTestAgentDispatch2}, got,
	)
}

func TestStringsx_Split_Ugly(t *core.T) {
	got := Split("abc", "")
	core.AssertEqual(
		t, []string{"a", "b", "c"}, got,
	)
}

func TestStringsx_SplitN_Good(t *core.T) {
	got := SplitN("key=value=extra", "=", 2)
	core.AssertEqual(
		t, []string{"key", "value=extra"}, got,
	)
}

func TestStringsx_SplitN_Bad(t *core.T) {
	got := SplitN("key=value", "=", 0)
	core.AssertNil(
		t, got,
	)
}

func TestStringsx_SplitN_Ugly(t *core.T) {
	got := SplitN("a=b=c", "=", -1)
	core.AssertEqual(
		t, []string{"a", "b", "c"}, got,
	)
}

func TestStringsx_SplitSeq_Good(t *core.T) {
	var got []string
	for part := range SplitSeq("agent,dispatch", ",") {
		got = append(got, part)
	}
	core.AssertEqual(t, []string{"agent", "dispatch"}, got)
}

func TestStringsx_SplitSeq_Bad(t *core.T) {
	var got []string
	for part := range SplitSeq("agent", ",") {
		got = append(got, part)
	}
	core.AssertEqual(t, []string{"agent"}, got)
}

func TestStringsx_SplitSeq_Ugly(t *core.T) {
	var got []string
	for part := range SplitSeq("abc", "") {
		got = append(got, part)
	}
	core.AssertEqual(t, []string{"a", "b", "c"}, got)
}

func TestStringsx_ToLower_Good(t *core.T) {
	got := ToLower("AGENT")
	core.AssertEqual(
		t, "agent", got,
	)
}

func TestStringsx_ToLower_Bad(t *core.T) {
	got := ToLower("agent-01")
	core.AssertEqual(
		t, "agent-01", got,
	)
}

func TestStringsx_ToLower_Ugly(t *core.T) {
	got := ToLower("")
	core.AssertEqual(
		t, "", got,
	)
}

func TestStringsx_ToUpper_Good(t *core.T) {
	got := ToUpper("agent")
	core.AssertEqual(
		t, "AGENT", got,
	)
}

func TestStringsx_ToUpper_Bad(t *core.T) {
	got := ToUpper("AGENT-01")
	core.AssertEqual(
		t, "AGENT-01", got,
	)
}

func TestStringsx_ToUpper_Ugly(t *core.T) {
	got := ToUpper("")
	core.AssertEqual(
		t, "", got,
	)
}

func TestStringsx_TrimPrefix_Good(t *core.T) {
	got := TrimPrefix(sonarStringsxTestAgentDispatch, "agent.")
	core.AssertEqual(
		t, "dispatch", got,
	)
}

func TestStringsx_TrimPrefix_Bad(t *core.T) {
	got := TrimPrefix(sonarStringsxTestAgentDispatch, "task.")
	core.AssertEqual(
		t, sonarStringsxTestAgentDispatch, got,
	)
}

func TestStringsx_TrimPrefix_Ugly(t *core.T) {
	got := TrimPrefix(sonarStringsxTestAgentDispatch, "")
	core.AssertEqual(
		t, sonarStringsxTestAgentDispatch, got,
	)
}

func TestStringsx_TrimSpace_Good(t *core.T) {
	got := TrimSpace("  agent  ")
	core.AssertEqual(
		t, "agent", got,
	)
}

func TestStringsx_TrimSpace_Bad(t *core.T) {
	got := TrimSpace("agent")
	core.AssertEqual(
		t, "agent", got,
	)
}

func TestStringsx_TrimSpace_Ugly(t *core.T) {
	got := TrimSpace("\n\tagent\r\n")
	core.AssertEqual(
		t, "agent", got,
	)
}

func TestStringsx_TrimSuffix_Good(t *core.T) {
	got := TrimSuffix(sonarStringsxTestManifestYaml, ".yaml")
	core.AssertEqual(
		t, "manifest", got,
	)
}

func TestStringsx_TrimSuffix_Bad(t *core.T) {
	got := TrimSuffix(sonarStringsxTestManifestYaml, ".json")
	core.AssertEqual(
		t, sonarStringsxTestManifestYaml, got,
	)
}

func TestStringsx_TrimSuffix_Ugly(t *core.T) {
	got := TrimSuffix(sonarStringsxTestManifestYaml, "")
	core.AssertEqual(
		t, sonarStringsxTestManifestYaml, got,
	)
}
