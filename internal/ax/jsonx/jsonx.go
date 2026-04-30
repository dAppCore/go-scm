// SPDX-License-Identifier: EUPL-1.2

package jsonx

import (
	"encoding/json"
	"io"
)

func Marshal(v any) ([]byte, error) { return json.Marshal(v) }
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}
func NewDecoder(r io.Reader) *json.Decoder { return json.NewDecoder(r) }
func NewEncoder(w io.Writer) *json.Encoder { return json.NewEncoder(w) }
func Unmarshal(data []byte, v any) error   { return json.Unmarshal(data, v) }
