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

// Package runes provides some functions for unicode code points.
package runes

type pair struct{ low, high rune }

type subset []pair

func (s subset) in(r rune) bool {
	for _, p := range s {
		if r >= p.low && r <= p.high {
			return true
		}
	}
	return false
}

// IsScalar tests if the rune is a "scalar value" according to RFC9839.
func IsScalar(r rune) bool { return scalars.in(r) }

var scalars = subset{
	{0, 0xd7ff},
	{0xe000, 0x10ffff},
}

// IsXMLChar tests if the rune is a "XML character", according to RFC9839.
func IsXMLChar(r rune) bool { return xmlChars.in(r) }

var xmlChars = subset{
	{0x20, 0xd7ff},
	{0xa, 0xa},
	{0x9, 0x9},
	{0xd, 0xd},
	{0xe000, 0xfffd},
	{0x10000, 0x10ffff},
}

// IsAssignable tests if the rune is a "Unicode assignable" character, according to RFC9839.
func IsAssignable(r rune) bool { return assignables.in(r) }

var assignables = subset{
	{0x20, 0x7e},
	{0xa, 0xa},
	{0x9, 0x9},
	{0xd, 0xd},
	{0xa0, 0xd7ff},
	{0xe000, 0xfdcf},
	{0xfdf0, 0xfffd},
	{0x10000, 0x1fffd},
	{0x20000, 0x2fffd},
	{0x30000, 0x3fffd},
	{0x40000, 0x4fffd},
	{0x50000, 0x5fffd},
	{0x60000, 0x6fffd},
	{0x70000, 0x7fffd},
	{0x80000, 0x8fffd},
	{0x90000, 0x9fffd},
	{0xa0000, 0xafffd},
	{0xb0000, 0xbfffd},
	{0xc0000, 0xcfffd},
	{0xd0000, 0xdfffd},
	{0xe0000, 0xefffd},
	{0xf0000, 0xffffd},
	{0x100000, 0x10fffd},
}

// IsAttributeName tests if rune is a valid character for an HTML attribute.
// See HTML5 spec, section 13.1.2.3
// https://html.spec.whatwg.org/multipage/syntax.html#syntax-attributes
func IsAttributeName(r rune) bool { return attributeNames.in(r) }

var attributeNames = subset{
	{0x3f, 0x7e},
	{0xa0, 0xfdcf},
	{0x21, 0x21},
	{0x23, 0x26},
	{0x28, 0x2e},
	{0x30, 0x3c},
	{0xfdf0, 0xfffd},
	{0x10000, 0x1fffd},
	{0x20000, 0x2fffd},
	{0x30000, 0x3fffd},
	{0x40000, 0x4fffd},
	{0x50000, 0x5fffd},
	{0x60000, 0x6fffd},
	{0x70000, 0x7fffd},
	{0x80000, 0x8fffd},
	{0x90000, 0x9fffd},
	{0xa0000, 0xafffd},
	{0xb0000, 0xbfffd},
	{0xc0000, 0xcfffd},
	{0xd0000, 0xdfffd},
	{0xe0000, 0xefffd},
	{0xf0000, 0xffffd},
	{0x100000, 0x10fffd},
}
