// SPDX-License-Identifier: EUPL-1.2

package jsonx

import (
	"math"

	core "dappco.re/go"
)

func TestJsonx_Marshal_Good(t *core.T) {
	got, err := Marshal(map[string]string{"agent": "codex"})
	core.AssertNoError(t, err)
	core.AssertContains(t, string(got), "codex")
}

func TestJsonx_Marshal_Bad(t *core.T) {
	_, err := Marshal(math.Inf(1))
	core.AssertError(
		t, err,
	)
}

func TestJsonx_Marshal_Ugly(t *core.T) {
	got, err := Marshal(nil)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "null", string(got))
}

func TestJsonx_MarshalIndent_Good(t *core.T) {
	got, err := MarshalIndent(map[string]string{"agent": "codex"}, "", "  ")
	core.AssertNoError(t, err)
	core.AssertContains(t, string(got), "\n")
}

func TestJsonx_MarshalIndent_Bad(t *core.T) {
	_, err := MarshalIndent(make(chan int), "", "  ")
	core.AssertError(
		t, err,
	)
}

func TestJsonx_MarshalIndent_Ugly(t *core.T) {
	got, err := MarshalIndent([]string{}, "", "  ")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "[]", string(got))
}

func TestJsonx_NewDecoder_Good(t *core.T) {
	decoder := NewDecoder(core.NewReader(`{"agent":"codex"}`))
	var got map[string]string
	err := decoder.Decode(&got)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "codex", got["agent"])
}

func TestJsonx_NewDecoder_Bad(t *core.T) {
	decoder := NewDecoder(core.NewReader(`{`))
	var got map[string]string
	err := decoder.Decode(&got)
	core.AssertError(t, err)
}

func TestJsonx_NewDecoder_Ugly(t *core.T) {
	decoder := NewDecoder(core.NewReader(""))
	var got map[string]string
	err := decoder.Decode(&got)
	core.AssertErrorIs(t, err, core.EOF)
}

func TestJsonx_NewEncoder_Good(t *core.T) {
	builder := core.NewBuilder()
	encoder := NewEncoder(builder)
	err := encoder.Encode(map[string]string{"agent": "codex"})
	core.AssertNoError(t, err)
	core.AssertContains(t, builder.String(), "codex")
}

func TestJsonx_NewEncoder_Bad(t *core.T) {
	builder := core.NewBuilder()
	encoder := NewEncoder(builder)
	err := encoder.Encode(math.Inf(1))
	core.AssertError(t, err)
}

func TestJsonx_NewEncoder_Ugly(t *core.T) {
	builder := core.NewBuilder()
	encoder := NewEncoder(builder)
	err := encoder.Encode(nil)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "null\n", builder.String())
}

func TestJsonx_Unmarshal_Good(t *core.T) {
	var got map[string]string
	err := Unmarshal([]byte(`{"agent":"codex"}`), &got)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "codex", got["agent"])
}

func TestJsonx_Unmarshal_Bad(t *core.T) {
	var got map[string]string
	err := Unmarshal([]byte(`{`), &got)
	core.AssertError(t, err)
}

func TestJsonx_Unmarshal_Ugly(t *core.T) {
	var got any
	err := Unmarshal([]byte(`null`), &got)
	core.AssertNoError(t, err)
	core.AssertNil(t, got)
}
