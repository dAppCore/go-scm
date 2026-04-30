// SPDX-License-Identifier: EUPL-1.2

package osx

import (
	"os"

	core "dappco.re/go"
)

const (
	sonarOsxTestBadPath  = "bad\x00path"
	sonarOsxTestCoreJson = "core.json"
)

func TestOsx_Getenv_Good(t *core.T) {
	t.Setenv("SCM_AX7_ENV", "ready")
	got := Getenv("SCM_AX7_ENV")
	core.AssertEqual(t, "ready", got)
}

func TestOsx_Getenv_Bad(t *core.T) {
	got := Getenv("SCM_AX7_MISSING")
	core.AssertEqual(
		t, "", got,
	)
}

func TestOsx_Getenv_Ugly(t *core.T) {
	got := Getenv("")
	core.AssertEqual(
		t, "", got,
	)
}

func TestOsx_Getwd_Good(t *core.T) {
	got, err := Getwd()
	core.AssertNoError(t, err)
	core.AssertTrue(t, got != "")
}

func TestOsx_Getwd_Bad(t *core.T) {
	got, err := Getwd()
	core.AssertNoError(t, err)
	core.AssertTrue(t, got != "/path/that/should/not/be/cwd")
}

func TestOsx_Getwd_Ugly(t *core.T) {
	core.AssertNotPanics(t, func() {
		_, _ = Getwd()
	})
}

func TestOsx_IsNotExist_Good(t *core.T) {
	_, err := Stat(core.Path(t.TempDir(), "missing"))
	got := IsNotExist(err)
	core.AssertTrue(t, got)
}

func TestOsx_IsNotExist_Bad(t *core.T) {
	_, err := Stat(t.TempDir())
	got := IsNotExist(err)
	core.AssertFalse(t, got)
}

func TestOsx_IsNotExist_Ugly(t *core.T) {
	got := IsNotExist(nil)
	core.AssertFalse(
		t, got,
	)
}

func TestOsx_MkdirAll_Good(t *core.T) {
	path := core.Path(t.TempDir(), "a", "b")
	err := MkdirAll(path, 0o755)
	core.AssertNoError(t, err)
	info, statErr := Stat(path)
	core.RequireNoError(t, statErr)
	core.AssertTrue(t, info.IsDir())
}

func TestOsx_MkdirAll_Bad(t *core.T) {
	file := core.Path(t.TempDir(), "file")
	core.RequireNoError(t, WriteFile(file, []byte("x"), 0o600))
	err := MkdirAll(core.Path(file, "child"), 0o755)
	core.AssertError(t, err)
}

func TestOsx_MkdirAll_Ugly(t *core.T) {
	err := MkdirAll(sonarOsxTestBadPath, 0o755)
	core.AssertError(
		t, err,
	)
}

func TestOsx_Open_Good(t *core.T) {
	path := core.Path(t.TempDir(), sonarOsxTestCoreJson)
	core.RequireNoError(t, WriteFile(path, []byte("ok"), 0o600))
	file, err := Open(path)
	core.RequireNoError(t, err)
	defer func() { core.AssertNoError(t, file.Close()) }()
	info, statErr := file.Stat()
	core.RequireNoError(t, statErr)
	core.AssertEqual(t, sonarOsxTestCoreJson, info.Name())
}

func TestOsx_Open_Bad(t *core.T) {
	_, err := Open(core.Path(t.TempDir(), "missing"))
	core.AssertError(
		t, err,
	)
}

func TestOsx_Open_Ugly(t *core.T) {
	_, err := Open(sonarOsxTestBadPath)
	core.AssertError(
		t, err,
	)
}

func TestOsx_OpenFile_Good(t *core.T) {
	path := core.Path(t.TempDir(), sonarOsxTestCoreJson)
	file, err := OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	core.RequireNoError(t, err)
	defer func() { core.AssertNoError(t, file.Close()) }()
	_, err = file.Write([]byte("ok"))
	core.AssertNoError(t, err)
}

func TestOsx_OpenFile_Bad(t *core.T) {
	_, err := OpenFile(core.Path(t.TempDir(), "missing"), os.O_RDONLY, 0)
	core.AssertError(
		t, err,
	)
}

func TestOsx_OpenFile_Ugly(t *core.T) {
	_, err := OpenFile(sonarOsxTestBadPath, os.O_RDONLY, 0)
	core.AssertError(
		t, err,
	)
}

func TestOsx_ReadDir_Good(t *core.T) {
	dir := t.TempDir()
	core.RequireNoError(t, WriteFile(core.Path(dir, sonarOsxTestCoreJson), []byte("ok"), 0o600))
	entries, err := ReadDir(dir)
	core.AssertNoError(t, err)
	core.AssertLen(t, entries, 1)
}

func TestOsx_ReadDir_Bad(t *core.T) {
	_, err := ReadDir(core.Path(t.TempDir(), "missing"))
	core.AssertError(
		t, err,
	)
}

func TestOsx_ReadDir_Ugly(t *core.T) {
	_, err := ReadDir(sonarOsxTestBadPath)
	core.AssertError(
		t, err,
	)
}

func TestOsx_ReadFile_Good(t *core.T) {
	path := core.Path(t.TempDir(), sonarOsxTestCoreJson)
	core.RequireNoError(t, WriteFile(path, []byte("ok"), 0o600))
	got, err := ReadFile(path)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "ok", string(got))
}

func TestOsx_ReadFile_Bad(t *core.T) {
	_, err := ReadFile(core.Path(t.TempDir(), "missing"))
	core.AssertError(
		t, err,
	)
}

func TestOsx_ReadFile_Ugly(t *core.T) {
	_, err := ReadFile(sonarOsxTestBadPath)
	core.AssertError(
		t, err,
	)
}

func TestOsx_Stat_Good(t *core.T) {
	path := core.Path(t.TempDir(), sonarOsxTestCoreJson)
	core.RequireNoError(t, WriteFile(path, []byte("ok"), 0o600))
	info, err := Stat(path)
	core.AssertNoError(t, err)
	core.AssertEqual(t, sonarOsxTestCoreJson, info.Name())
}

func TestOsx_Stat_Bad(t *core.T) {
	_, err := Stat(core.Path(t.TempDir(), "missing"))
	core.AssertError(
		t, err,
	)
}

func TestOsx_Stat_Ugly(t *core.T) {
	_, err := Stat(sonarOsxTestBadPath)
	core.AssertError(
		t, err,
	)
}

func TestOsx_UserHomeDir_Good(t *core.T) {
	got, err := UserHomeDir()
	core.AssertNoError(t, err)
	core.AssertTrue(t, got != "")
}

func TestOsx_UserHomeDir_Bad(t *core.T) {
	got, err := UserHomeDir()
	core.AssertNoError(t, err)
	core.AssertTrue(t, got != "/definitely/not/home")
}

func TestOsx_UserHomeDir_Ugly(t *core.T) {
	core.AssertNotPanics(t, func() {
		_, _ = UserHomeDir()
	})
}

func TestOsx_WriteFile_Good(t *core.T) {
	path := core.Path(t.TempDir(), sonarOsxTestCoreJson)
	err := WriteFile(path, []byte("ok"), 0o600)
	core.AssertNoError(t, err)
	got, readErr := ReadFile(path)
	core.RequireNoError(t, readErr)
	core.AssertEqual(t, "ok", string(got))
}

func TestOsx_WriteFile_Bad(t *core.T) {
	err := WriteFile(core.Path(t.TempDir(), "missing", sonarOsxTestCoreJson), []byte("ok"), 0o600)
	core.AssertError(
		t, err,
	)
}

func TestOsx_WriteFile_Ugly(t *core.T) {
	err := WriteFile(sonarOsxTestBadPath, []byte("ok"), 0o600)
	core.AssertError(
		t, err,
	)
}
