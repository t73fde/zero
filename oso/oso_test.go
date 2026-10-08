//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package oso_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"t73f.de/r/zero/oso"
)

func readFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertContent(t *testing.T, name string, want []byte) {
	t.Helper()
	if got := readFile(t, name); !bytes.Equal(got, want) {
		t.Errorf("content of %q: expected %q, but got %q", name, want, got)
	}
}

func assertEntries(t *testing.T, dir string, want ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.Name())
	}
	if !slices.Equal(got, want) {
		t.Errorf("directory entries: expected %v, but got %v", want, got)
	}
}

func TestWriteString(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	fname := filepath.Join(dir, "happy")
	f, err := oso.SafeWrite(fname)
	if err != nil {
		t.Fatal("new", err)
	}
	defer f.RollbackIfNeeded()

	const content = "Hello OSO"
	n, err := f.WriteString(content)
	if err != nil {
		t.Fatal("write", err)
	}
	if n != len(content) {
		t.Errorf("written bytes, expected: %d, but got %d", len(content), n)
	}
	if err = f.Close(); err != nil {
		t.Fatal("close", err)
	}
	assertContent(t, fname, []byte(content))
	assertEntries(t, dir, "happy")
}

func TestWriteAndReadFrom(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("oso.go")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		write func(f *oso.File) error
	}{
		{"Write", func(f *oso.File) error { _, errW := f.Write(src); return errW }},
		{"ReadFrom", func(f *oso.File) error { _, errRF := f.ReadFrom(bytes.NewReader(src)); return errRF }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fname := filepath.Join(t.TempDir(), "copy")
			f, errT := oso.SafeWrite(fname)
			if errT != nil {
				t.Fatal("new", errT)
			}
			defer f.RollbackIfNeeded()
			if errT = tc.write(f); errT != nil {
				t.Fatal("write", errT)
			}
			if errT = f.Close(); errT != nil {
				t.Fatal("close", errT)
			}
			assertContent(t, fname, src)
		})
	}
}

func TestEmptyFile(t *testing.T) {
	t.Parallel()
	fname := filepath.Join(t.TempDir(), "empty")
	f, err := oso.SafeWrite(fname)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	assertContent(t, fname, nil)
}

func TestOverwrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	fname := filepath.Join(dir, "target")
	if err := os.WriteFile(fname, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := oso.SafeWrite(fname)
	if err != nil {
		t.Fatal(err)
	}
	defer f.RollbackIfNeeded()
	if _, err = f.WriteString("new"); err != nil {
		t.Fatal(err)
	}
	assertContent(t, fname, []byte("old")) // not visible before Close
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	assertContent(t, fname, []byte("new"))
	assertEntries(t, dir, "target")
}

func TestRollback(t *testing.T) {
	t.Parallel()
	t.Run("existing destination is kept", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		fname := filepath.Join(dir, "target")
		if err := os.WriteFile(fname, []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
		f, err := oso.SafeWrite(fname)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.WriteString("new"); err != nil {
			t.Fatal(err)
		}
		f.RollbackIfNeeded()
		if err = f.Close(); !errors.Is(err, oso.ErrRollback) {
			t.Errorf("close after rollback: expected %v, but got %v", oso.ErrRollback, err)
		}
		assertContent(t, fname, []byte("old"))
		assertEntries(t, dir, "target")
	})
	t.Run("no destination is created", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		f, err := oso.SafeWrite(filepath.Join(dir, "target"))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = f.WriteString("data")
		f.RollbackIfNeeded()
		assertEntries(t, dir)
	})
	t.Run("deferred rollback on early return", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		func() {
			f, err := oso.SafeWrite(filepath.Join(dir, "target"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.RollbackIfNeeded()
			_, _ = f.WriteString("data")
		}()
		assertEntries(t, dir)
	})
	t.Run("no effect after successful close", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		fname := filepath.Join(dir, "target")
		f, err := oso.SafeWrite(fname)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = f.WriteString("data")
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
		f.RollbackIfNeeded()
		assertContent(t, fname, []byte("data"))
	})
	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var f *oso.File
		f.RollbackIfNeeded()
	})
}

func TestCloseTwice(t *testing.T) {
	t.Parallel()
	f, err := oso.SafeWrite(filepath.Join(t.TempDir(), "target"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal("first close", err)
	}
	if err = f.Close(); err != nil {
		t.Error("second close", err)
	}
}

func TestWriteAfterClose(t *testing.T) {
	t.Parallel()
	fname := filepath.Join(t.TempDir(), "target")
	f, err := oso.SafeWrite(fname)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("data")
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString("more"); !errors.Is(err, os.ErrClosed) {
		t.Errorf("write after close: expected %v, but got %v", os.ErrClosed, err)
	}
	if err = f.Close(); err != nil {
		t.Errorf("close after rejected write: %v", err)
	}
	assertContent(t, fname, []byte("data"))
}

func TestInvalidPath(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"", ".", "..", "a/..", string(filepath.Separator)} {
		t.Run("path="+path, func(t *testing.T) {
			t.Parallel()
			f, err := oso.SafeWrite(path)
			if err == nil {
				f.RollbackIfNeeded()
				t.Fatal("expected error")
			}
			if !errors.Is(err, os.ErrInvalid) {
				t.Errorf("expected %v, but got %v", os.ErrInvalid, err)
			}
		})
	}
}

func TestMissingDirectory(t *testing.T) {
	t.Parallel()
	f, err := oso.SafeWrite(filepath.Join(t.TempDir(), "missing", "target"))
	if err == nil {
		f.RollbackIfNeeded()
		t.Fatal("expected error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected %v, but got %v", os.ErrNotExist, err)
	}
	var pe *os.PathError
	if !errors.As(err, &pe) || pe.Op != "new" {
		t.Errorf("expected *PathError with op \"new\", but got %#v", err)
	}
}

func TestDestinationIsDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "target")
	if err := os.Mkdir(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := oso.SafeWrite(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer f.RollbackIfNeeded()
	_, _ = f.WriteString("data")
	if err = f.Close(); err == nil {
		t.Fatal("expected error on close")
	}
	assertEntries(t, dir, "target") // temporary file removed
}

func TestRelativePath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(wd, filepath.Join(dir, "target"))
	if err != nil {
		t.Skip("no relative path to temporary directory:", err)
	}
	f, err := oso.SafeWrite(rel)
	if err != nil {
		t.Fatal(err)
	}
	defer f.RollbackIfNeeded()
	_, _ = f.WriteString("data")
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	assertContent(t, filepath.Join(dir, "target"), []byte("data"))
}

func TestPrefix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f, err := oso.SafeWriteWith(filepath.Join(dir, "target"), "tmp-*.part")
	if err != nil {
		t.Fatal(err)
	}
	defer f.RollbackIfNeeded()
	matches, err := filepath.Glob(filepath.Join(dir, "tmp-*.part"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("expected one temporary file, got %v (err: %v)", matches, err)
	}
	if strings.Contains(filepath.Base(matches[0]), "*") {
		t.Errorf("pattern not expanded: %q", matches[0])
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	assertEntries(t, dir, "target")
}

func TestFileMode(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not meaningful on Windows")
	}
	fname := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(fname, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := oso.SafeWrite(fname)
	if err != nil {
		t.Fatal(err)
	}
	defer f.RollbackIfNeeded()
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(fname)
	if err != nil {
		t.Fatal(err)
	}
	if got := fi.Mode().Perm(); got != 0o600 {
		t.Errorf("file mode: expected %v, but got %v", os.FileMode(0o600), got)
	}
}
