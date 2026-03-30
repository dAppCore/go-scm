// SPDX-License-Identifier: EUPL-1.2

package jsonx

import (
	"io"

	json "github.com/goccy/go-json"
)

// Marshal mirrors encoding/json.Marshal.
//
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent mirrors encoding/json.MarshalIndent.
//
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// NewDecoder mirrors encoding/json.NewDecoder.
//
func NewDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder mirrors encoding/json.NewEncoder.
//
func NewEncoder(w io.Writer) *json.Encoder {
	return json.NewEncoder(w)
}

// Unmarshal mirrors encoding/json.Unmarshal.
//
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
