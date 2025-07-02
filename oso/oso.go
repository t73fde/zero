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

// Package oso provides safe, atomic file writing.
//
// Writing to a file is easy, if you ignore all possible error conditions.
//
// Writing to a file is hard. You must handle errors returned by write or close
// operations. Even if you successfully close the file, it might not be written
// to disk. You need to synchronize the file with the operating system. Even
// that does not guarantee that the file content is physically stored on disk.
// What happens if multiple processes write to the same file simultaneously?
// What happens to the previous content of a file if an error occurs during a
// write operation?
//
// To handle all these issues, write everything to a temporary file first.
// After closing the temporary file, rename it to the destination file. On most
// file systems, this approximates an atomic operation.
//
// Use it this way:
//
//	func writeData(filename string, data []byte) error {
//	        f, err := oso.SafeWrite(filename)
//	        if err != nil {
//	                return err
//	        }
//	        defer f.RollbackIfNeeded()
//
//	        _, _ = f.Write(data)
//	        return f.Close()
//	}
//
// The package is named after the manufacturer of some safes owned by Scrooge McDuck.
package oso

import (
	"cmp"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Roughly based on https://github.com/kjk/common/blob/main/atomicfile

// File allows to write content in a safe / atomic matter to a file.
type File struct {
	path string
	dir  string
	tmpf *os.File
	err  error
}

var (
	// Ensure some interfaces.
	_ io.WriteCloser  = &File{}
	_ io.ReaderFrom   = &File{}
	_ io.StringWriter = &File{}
)

// SafeWrite creates a new file with the given path.
func SafeWrite(path string) (*File, error) { return SafeWriteWith(path, "") }

// SafeWriteWith creates a new file with the given path and prefix for the
// temporary file.
func SafeWriteWith(path, prefix string) (*File, error) {
	path = filepath.Clean(path)
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, &fs.PathError{Op: "new", Path: path, Err: err}
	}
	dir, tmpname := filepath.Split(path)
	if prefix != "" {
		tmpname = prefix
	}
	if tmpname == "" || tmpname == "." || tmpname == ".." {
		return nil, &fs.PathError{Op: "new", Path: path, Err: os.ErrInvalid}
	}
	if dir == "" {
		dir = "."
	}
	if prefix != "" {
		tmpname = prefix
	}
	tmpf, err := os.CreateTemp(dir, tmpname)
	if err != nil {
		return nil, &fs.PathError{Op: "new", Path: path, Err: err}
	}
	return &File{
		path: path,
		dir:  dir,
		tmpf: tmpf,
	}, nil
}

// ----- io.WriteCloser methods

// Write some bytes to the file.
func (f *File) Write(b []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	n, err := f.tmpf.Write(b)
	return n, f.processError(err)
}

// Close the file, make all changes visible.
func (f *File) Close() error {
	// Here happens all the magic of atomicity

	if f.tmpf == nil { // already closed
		return f.err
	}

	// TODO: need a mutex for concurrent access?
	tmpf := f.tmpf
	f.tmpf = nil

	// Auto-rollback if something happens: delete temp file
	disableRollback := false
	defer func() {
		if !disableRollback {
			_ = os.Remove(f.tmpf.Name()) // Ignore error, just do your best
		}
	}()

	// Try to do the best by trying to sync and close.
	errSync := tmpf.Sync()   // First Sync, then Close
	errClose := tmpf.Close() // Must be done to allow to remove file in rollback

	if f.err != nil {
		return f.err
	}

	err := cmp.Or(errSync, errClose)
	if err == nil {
		// os.Rename will remove possibly existing file
		if err = os.Rename(tmpf.Name(), f.path); err == nil {
			disableRollback = true
		}

		// Give OS some hint to sync directory b/c storage of metadata.
		if dirf, errDir := os.Open(f.dir); errDir == nil && dirf != nil {
			_ = dirf.Sync()
			_ = dirf.Close()
		}
	}

	if f.err == nil {
		f.err = err
	}
	return f.err
}

// ----- optimizing methods

// WriteString writes a string to the file.
//
// Optimizes io.WriteString(w, s) by implementing io.StringWriter.
func (f *File) WriteString(s string) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	n, err := f.tmpf.WriteString(s)
	return n, f.processError(err)
}

// ReadFrom reads data from a source and writes it to the file.
//
// Implements io.ReaderFrom, to optimize io.Copy.
func (f *File) ReadFrom(r io.Reader) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	n, err := f.tmpf.ReadFrom(r)
	return n, f.processError(err)
}

// ----- utitlity functions

// RollbackIfNeeded everything to initial state of file system.
//
// Should be used as a defer function when working with oso.File.
func (f *File) RollbackIfNeeded() {
	if f == nil || f.tmpf == nil {
		// no file or file already closed: do nothing.
		return
	}
	f.err = ErrRollback
	_ = f.Close()
}

// ErrRollback signals that the File was rolled back.
var ErrRollback = errors.New("rollback")

func (f *File) processError(err error) error {
	if err != nil {
		if f.err == nil {
			f.err = err
		}
		_ = f.Close()
	}
	return err
}
