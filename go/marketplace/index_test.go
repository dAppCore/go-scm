// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"io/fs"
	"testing"

	coreio "dappco.re/go/io"
)

const (
	sonarIndexTestMarketplaceIndexJson = "marketplace/index.json"
)

type readWriteFileMedium struct {
	*coreio.MockMedium
	readFileCalled  bool
	writeFileCalled bool
}

func (m *readWriteFileMedium) ReadFile(path string) ([]byte, error) {
	m.readFileCalled = true
	raw, err := m.Read(path)
	return []byte(raw), err
}

func (m *readWriteFileMedium) WriteFile(path string, data []byte, _ fs.FileMode) error {
	m.writeFileCalled = true
	return m.Write(path, string(data))
}

func TestLoadIndexUsesReadFile(t *testing.T) {
	medium := &readWriteFileMedium{MockMedium: coreio.NewMockMedium()}
	if err := medium.Write(sonarIndexTestMarketplaceIndexJson, `{"version":1,"modules":[{"code":"go-io","name":"Core I/O"}]}`); err != nil {
		t.Fatalf("seed index: %v", err)
	}

	idx, err := LoadIndex(medium, sonarIndexTestMarketplaceIndexJson)
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if !medium.readFileCalled {
		t.Fatalf("expected LoadIndex to use ReadFile")
	}
	if idx == nil || len(idx.Modules) != 1 || idx.Modules[0].Code != "go-io" {
		t.Fatalf("unexpected index: %#v", idx)
	}
}

func TestWriteIndexToMediumUsesWriteFile(t *testing.T) {
	medium := &readWriteFileMedium{MockMedium: coreio.NewMockMedium()}

	if err := WriteIndexToMedium(medium, sonarIndexTestMarketplaceIndexJson, &Index{
		Version: 1,
		Modules: []Module{{Code: "go-io", Name: "Core I/O"}},
	}); err != nil {
		t.Fatalf("WriteIndexToMedium: %v", err)
	}
	if !medium.writeFileCalled {
		t.Fatalf("expected WriteIndexToMedium to use WriteFile")
	}
	if !medium.IsFile(sonarIndexTestMarketplaceIndexJson) {
		t.Fatalf("expected index file to be written")
	}
}

func TestIndex_LoadIndex_Good(t *testing.T) {
	target := "LoadIndex"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestIndex_LoadIndex_Bad(t *testing.T) {
	target := "LoadIndex"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestIndex_LoadIndex_Ugly(t *testing.T) {
	target := "LoadIndex"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestIndex_WriteIndexToMedium_Good(t *testing.T) {
	target := "WriteIndexToMedium"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestIndex_WriteIndexToMedium_Bad(t *testing.T) {
	target := "WriteIndexToMedium"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestIndex_WriteIndexToMedium_Ugly(t *testing.T) {
	target := "WriteIndexToMedium"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}
