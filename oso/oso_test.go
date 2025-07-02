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
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestHappy(t *testing.T) {
	const fname = "happy"
	f, err := New(fname)
	defer f.RollbackIfNeeded()
	if err != nil {
		t.Error("new", err)
		return
	}
	const content = "Hello OSO"
	n, err := f.WriteString(content)
	if err != nil {
		t.Error("write", err)
		return
	}
	if n != len(content) {
		t.Errorf("content-length, expected: %d, but got %d", len(content), n)
	}
	err = f.Close()
	if err != nil {
		t.Error("close", err)
		return
	}

	assertFileContent(t, fname, []byte(content))
	_ = os.Remove(fname)
}

func TestCopyWrite(t *testing.T) {
	content, err := getFileData("oso.go")
	if err != nil {
		panic(err)
	}

	const fname = "oso.go.copy"
	f, err := New(fname)
	defer f.RollbackIfNeeded()
	if err != nil {
		t.Error("new", err)
		return
	}
	_, err = f.Write(content)
	if err != nil {
		t.Error("write", err)
		return
	}
	err = f.Close()
	if err != nil {
		t.Error("close", err)
		return
	}
	assertFileContent(t, fname, content)
	_ = os.Remove(fname)
}

func TestCopyReadFrom(t *testing.T) {
	rf, err := os.Open("oso.go")
	if err != nil {
		panic(err)
	}

	const fname = "oso.go.copy"
	f, err := New(fname)
	defer f.RollbackIfNeeded()
	if err != nil {
		t.Error("new", err)
		return
	}
	_, err = f.ReadFrom(rf)
	if err != nil {
		t.Error("readFrom", err)
		return
	}
	err = f.Close()
	if err != nil {
		t.Error("close", err)
		return
	}

	content, err := getFileData("oso.go")
	if err != nil {
		panic(err)
	}
	assertFileContent(t, fname, content)
	_ = os.Remove(fname)
}

func assertFileContent(t *testing.T, fname string, content []byte) {
	t.Helper()
	data, err := getFileData(fname)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(content, data) {
		t.Errorf("expected content %q, but got %q", string(content), string(data))
	}
}

func getFileData(fname string) ([]byte, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	err = f.Close()
	if err != nil {
		return nil, fmt.Errorf("close: %w", err)
	}
	return data, err
}
