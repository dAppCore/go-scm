// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"errors"
	"io/fs"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

type mediumReadFile interface {
	ReadFile(string) ([]byte, error)
}

type mediumWriteFile interface {
	WriteFile(string, []byte, fs.FileMode) error
}

type mediumWriteFileBytes interface {
	WriteFile(string, []byte) error
}

type mediumWriteFileString interface {
	WriteFile(string, string) error
}

// LoadIndex reads a marketplace index through an io.Medium.
func LoadIndex(m coreio.Medium, path string) (*Index, error) {
	if m == nil {
		return nil, errors.New("marketplace.LoadIndex: medium is required")
	}
	raw, err := readMediumFile(m, path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Index{Version: 1, Modules: []Module{}}, nil
		}
		return nil, err
	}
	return ParseIndex(raw)
}

// WriteIndexToMedium writes a marketplace index through an io.Medium.
func WriteIndexToMedium(m coreio.Medium, path string, idx *Index) error {
	if m == nil {
		return errors.New("marketplace.WriteIndexToMedium: medium is required")
	}
	if idx == nil {
		return errors.New("marketplace.WriteIndexToMedium: index is required")
	}
	marshalResult := core.JSONMarshalIndent(idx, "", "  ")
	if !marshalResult.OK {
		return core.E("marketplace.WriteIndexToMedium", "encode index", nil)
	}
	return writeMediumFile(m, path, marshalResult.Value.([]byte))
}

func readMediumFile(m coreio.Medium, path string) ([]byte, error) {
	if m == nil {
		return nil, errors.New("marketplace.readMediumFile: medium is required")
	}
	if reader, ok := m.(mediumReadFile); ok {
		return reader.ReadFile(path)
	}
	raw, err := m.Read(path)
	if err != nil {
		return nil, err
	}
	return []byte(raw), nil
}

func writeMediumFile(m coreio.Medium, path string, data []byte) error {
	if m == nil {
		return errors.New("marketplace.writeMediumFile: medium is required")
	}
	if writer, ok := m.(mediumWriteFile); ok {
		return writer.WriteFile(path, data, 0o600)
	}
	if writer, ok := m.(mediumWriteFileBytes); ok {
		return writer.WriteFile(path, data)
	}
	if writer, ok := m.(mediumWriteFileString); ok {
		return writer.WriteFile(path, string(data))
	}
	return m.Write(path, string(data))
}
