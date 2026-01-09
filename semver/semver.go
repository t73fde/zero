// -----------------------------------------------------------------------------
// Copyright (c) 2026-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2026-present Detlef Stern
// -----------------------------------------------------------------------------

// Package semver implements creation, comparison, and more of semantic version data.
//
// The general form of a semantic version string is:
//
//	MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
//
// where square brackets indicate optional parts. MAJOR, MINMOR, and PATCH are
// decimal numbers without leading zeroes; PRERELEASE and BUILD are each one or
// more sequences of alphanumeric characters, separated by dots.
// PRERELEASE must not have leading zeroes.
//
// This package follows Semantic Versioning 2.0.0 (see semver.org).
package semver

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var reSemVer = regexp.MustCompile(
	`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// Parse a string as a semantic version string.
func Parse(s string) (SemVer, bool) {
	if m := reSemVer.FindStringSubmatch(s); len(m) > 0 {
		if major, errMajor := strconv.ParseUint(m[1], 10, 64); errMajor == nil {
			if minor, errMinor := strconv.ParseUint(m[2], 10, 64); errMinor == nil {
				if patch, errPatch := strconv.ParseUint(m[3], 10, 64); errPatch == nil {
					return SemVer{
						Major:      major,
						Minor:      minor,
						Patch:      patch,
						PreRelease: m[4],
						Build:      m[5],
					}, true
				}
			}
		}
	}
	return SemVer{}, false
}

// MustParse parses the string as a semantic version and panics if this is not possible.
func MustParse(s string) SemVer {
	v, ok := Parse(s)
	if !ok {
		panic(fmt.Sprintf("%q is not a valid SemVer string", s))
	}
	return v
}

// SemVer stores the parsed data for a semantic version string.
type SemVer struct {
	Major      uint64
	Minor      uint64
	Patch      uint64
	PreRelease string
	Build      string
}

// String returns a string representation of the semantic version.
func (v SemVer) String() string {
	var sb strings.Builder
	sb.WriteString(strconv.FormatUint(v.Major, 10))
	sb.WriteByte('.')
	sb.WriteString(strconv.FormatUint(v.Minor, 10))
	sb.WriteByte('.')
	sb.WriteString(strconv.FormatUint(v.Patch, 10))
	if p := v.PreRelease; p != "" {
		sb.WriteByte('-')
		sb.WriteString(p)
	}
	if b := v.Build; b != "" {
		sb.WriteByte('+')
		sb.WriteString(b)
	}
	return sb.String()
}

// Compare two semantic versions.
func (v SemVer) Compare(o SemVer) int {
	if c := cmp.Compare(v.Major, o.Major); c != 0 {
		return c
	}
	if c := cmp.Compare(v.Minor, o.Minor); c != 0 {
		return c
	}
	if c := cmp.Compare(v.Patch, o.Patch); c != 0 {
		return c
	}
	vp := v.PreRelease
	op := o.PreRelease
	if vp == "" && op != "" {
		return 1
	}
	if vp != "" && op == "" {
		return -1
	}
	vs := strings.Split(vp, ".")
	os := strings.Split(op, ".")
	return slices.CompareFunc(vs, os, func(ve, oe string) int {
		if vi, errV := strconv.ParseUint(ve, 10, 64); errV == nil {
			if oi, errO := strconv.ParseUint(oe, 10, 64); errO == nil {
				return cmp.Compare(vi, oi)
			}
		}
		return strings.Compare(ve, oe)
	})
}

// IncPatch increments the patch version.
func (v *SemVer) IncPatch() {
	v.Patch++
	v.PreRelease = ""
	v.Build = ""
}

// IncMinor increments the minor version.
func (v *SemVer) IncMinor() {
	v.Minor++
	v.Patch = 0
	v.PreRelease = ""
	v.Build = ""
}

// IncMajor increments the major version.
func (v *SemVer) IncMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0
	v.PreRelease = ""
	v.Build = ""
}
