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

package oso

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteErrorIsSticky(t *testing.T) {
	t.Parallel()
	ops := []struct {
		name string
		do   func(f *File) error
	}{
		{"Write", func(f *File) error { _, err := f.Write([]byte("x")); return err }},
		{"WriteString", func(f *File) error { _, err := f.WriteString("x"); return err }},
		{"ReadFrom", func(f *File) error { _, err := f.ReadFrom(strings.NewReader("x")); return err }},
	}
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			fname := filepath.Join(dir, "target")
			f, err := SafeWrite(fname)
			if err != nil {
				t.Fatal(err)
			}
			defer f.RollbackIfNeeded()

			_ = f.tmpf.Close() // force a write error

			first := op.do(f)
			if !errors.Is(first, os.ErrClosed) {
				t.Fatalf("expected %v, but got %v", os.ErrClosed, first)
			}
			if second := op.do(f); second != first {
				t.Errorf("error not sticky: expected %v, but got %v", first, second)
			}
			if err = f.Close(); err != first {
				t.Errorf("close: expected %v, but got %v", first, err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Errorf("expected empty directory, but got %v", entries)
			}
		})
	}
}
