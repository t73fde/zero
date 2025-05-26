// -----------------------------------------------------------------------------
// Copyright (c) 2023-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2023-present Detlef Stern
// -----------------------------------------------------------------------------

// Package snow provides a generic key to be used as an URI element, as a
// primary key in a database, or for other usage.
package snow

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Key is the generic, external primary key for all data.
//
// Snowflake/TSID:
// * 42 bit timestamp, enough to be used in the year 2160.
// * 22 bit application / sequence number
//   - 0-20 bit application defined data, e.g. for tables, nodes, ...
//   - 2-22 bit sequence number
//
// The timestamp value starts at 2024-06-01. 42 bits of milliseconds allow an
// end date in the year 2163.
//
// The rest of the 64 bits, ie 22 bits can be splitted at the users demand.
// Two bits are always reserved to be used as a sequence number if two keys are
// generated within the same millisecond. This can be enlarged up 22 bits.
// A maximum of 20 bits can be used for the application. The more bits are
// used, the less bits are available for the sequence number. An application
// can use the bits to store the number of a database table, or the number of a
// computing node.
type Key uint64

// Invalid is the default invalid key.
const Invalid Key = 0

const (
	timestampBits = 42
	randomBits    = 22

	maxTimeStamp = 1<<timestampBits - 1
)

// MaxAppBits states the maximum number of bits reserved for the application
// defined part of the key.
const MaxAppBits = 20

// Parse will parse a string into an external key.
func Parse(s string) (Key, error) {
	result := Key(0)
	for i := range len(s) {
		ch := s[i]
		if ch == '-' && i > 0 && i < len(s)-1 {
			continue
		}
		if '0' <= ch && ch <= 128 {
			val := decode32map[ch-'0']
			if 0 <= val && val <= 31 {
				if result&0xF800000000000000 != 0 {
					return Invalid, fmt.Errorf("does not fit in uint64: %q / %x", s, uint64(result))
				}
				result = (result << 5) | Key(val)
				continue
			}
		}
		return result, fmt.Errorf("non base-32 character %c/%v found", ch, ch)
	}
	return result, nil
}

// MustParse parses the string into an external key, and panics if that is not possible.
func MustParse(s string) Key {
	key, err := Parse(s)
	if err == nil {
		return key
	}
	panic(err)
}

// IsInvalid returns true if the key is definitely an invalid key.
func (key Key) IsInvalid() bool { return key == Invalid }

// IsValid returns true if the key is definitely an invalid key.
func (key Key) IsValid() bool { return key != Invalid }

var decode32map = [...]int8{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, -1, -1, -1, -1, -1, -1, // 0x30 .. 0x3f
	-1, 10, 11, 12, 13, 14, 15, 16, 17, 1, 18, 19, 1, 20, 21, 0, // 0x40 .. 0x4f
	22, 23, 24, 25, 26, 36, 27, 28, 29, 30, 31, -1, -1, -1, -1, -1, // 0x50 .. 0x5f
	-1, 10, 11, 12, 13, 14, 15, 16, 17, 1, 18, 19, 1, 20, 21, 0, // 0x60 .. 0x6f
	22, 23, 24, 25, 26, -1, 27, 28, 29, 30, 31, -1, -1, -1, -1, -1, // 0x70 .. 0x7f
}

// Time returns the timestamp value, when the key was generated.
func (key Key) Time() time.Time {
	return time.UnixMilli(int64(key>>randomBits) + epochAdjust)
}

// String returns a base-32 representation of the key as a string.
// It contains at most 13 characters.
func (key Key) String() string {
	if key == 0 {
		return "0"
	}
	temp, tpos := key.reverseEncode()

	var result [13]byte
	for i := range tpos {
		result[i] = temp[tpos-i-1]
	}
	return string(result[:tpos])
}

var sepMask = []uint16{
	0b0000000000000, // 0  = "ABCDEFGHJKMNP" (sentinel)
	0b0111111111111, // 1  = "A-B-C-D-E-F-G-H-J-K-M-N-P"
	0b0010101010101, // 2  = "A-BC-DE-FG-HJ-KM-NP"
	0b0001001001001, // 3  = "A-BCD-EFG-HJK-MNP"
	0b0000100010001, // 4  = "A-BCDE-FGHJ-KMNP"
	0b0000010000100, // 5  = "ABC-DEFGH-JKMNP"
	0b0000001000001, // 6  = "A-BCDEFG-HJKMNP"
	0b0000000100000, // 7  = "ABCDEF-GHJKMNP"
	0b0000000010000, // 8  = "ABCDE-FGHJKMNP"
	0b0000000001000, // 9  = "ABCD-EFGHJKMNP"
	0b0000000000100, // 10 = "ABC-DEFGHJKMNP"
	0b0000000000010, // 11 = "AB-CDEFGHJKMNP"
	0b0000000000001, // 12 = "A-BCDEFGHJKMNP"
}

// Format returns a string representing the key, where groups of key digits
// (base-32) are separated by a string. In contrast to String, all digits are
// returned, even leading zeroes.
//
// For example: Invalid.Format(4, "-") == "0-0000-0000-0000".
//
// If the separator contains base-32 digit characters, you will not be able to
// parse the result later.
//
// For example: Invalid.Format(4, "3") == "0300003000030000". Parsing it will
// result in an error (string too long), but you will not be able to shorten
// is correctly, in the general case.
//
// A group size less than one is interpreted as a group size of one. If the
// spearator is the emptay string, or if group size is greater than 12, no
// separators are included, returning just the 13 base-32 digits of the key.
//
// For example: Invalid.Format(4, "") == "0000000000000", and
// Invalid.Format(13, "-") == "0000000000000".
//
// If you want to parse a formatted key, use the standard library strings
// package: snow.Parse(strings.Join(strings.Split(key.Format(4, sep), sep), "")).
func (key Key) Format(groupSize int, sep string) string {
	if groupSize <= 0 {
		groupSize = 1
	}
	temp, tpos := key.reverseEncode()
	for ; tpos < len(temp); tpos++ {
		temp[tpos] = '0'
	}

	if sep == "" || groupSize >= len(temp) {
		var result [13]byte
		for i := range tpos {
			result[i] = temp[tpos-i-1]
		}
		return string(result[:tpos])
	}

	mask := sepMask[groupSize]
	var sb strings.Builder
	for tpos > 0 {
		tpos--
		_ = sb.WriteByte(temp[tpos])
		if mask%2 == 1 {
			_, _ = sb.WriteString(sep)
		}
		mask /= 2
	}
	return sb.String()
}

func (key Key) reverseEncode() ([13]byte, int) {
	u64 := uint64(key)
	temp := [13]byte{}
	tpos := 0
	for u64 > 0 {
		temp[tpos] = base32chars[u64%32]
		tpos++
		u64 = u64 >> 5
	}
	return temp, tpos
}

const base32chars = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// Generator is a generator for unique keys as int64.
type Generator struct {
	mx      sync.Mutex // Protects the next two fields
	lastTS  uint64     // Last timestamp
	nextSeq uint64     // Next sequence number for lastTS
	appBits uint       // number of bits for application use. range: 0-MaxAppBits
	appMax  uint       // 1 << appBits (if appBits > 0; else: 0)
}

// New creates a new key generator with a given number of bits for
// application use.
func New(appBits uint) *Generator {
	if appBits > MaxAppBits {
		panic(fmt.Sprintf("key generator need too many bits (max %d): %v", appBits, MaxAppBits))
	}
	return &Generator{
		appBits: appBits,
		appMax:  1 << appBits,
	}
}

// epochAdjust is used to make the timestamp values smaller, so they better fit
// in 42 bits.
//
// Its value is time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
const epochAdjust = 1717200000000

// Create generates a new key with the given application data.
func (gen *Generator) Create(appID uint) Key {
	if appID > 0 && appID >= gen.appMax {
		panic(fmt.Errorf("application value out of range: %v (max: %v)", appID, gen.appMax))
	}
	for {
		milli := uint64(time.Now().UnixMilli())
		var seq uint64

		gen.mx.Lock()
		if milli > gen.lastTS {
			gen.lastTS = milli
			gen.nextSeq = 1
			seq = 0
		} else {
			seq = gen.nextSeq
			gen.nextSeq++
		}
		gen.mx.Unlock()

		if seq < (1 << (randomBits - gen.appBits)) {
			ts := milli - epochAdjust
			if ts > maxTimeStamp {
				panic(fmt.Sprintf("timestamp %v exceeds largest possible value %v", ts, maxTimeStamp))
			}

			// 42bit=ts, kg.intBits=appId, 22-kg.intBits=seq
			k := (ts << randomBits) | (uint64(appID) << (randomBits - gen.appBits)) | seq
			return Key(k)
		}

		time.Sleep(1 * time.Millisecond)
	}
}

// AppID returns the application defined part of the key.
func (gen *Generator) AppID(key Key) uint {
	if appBits := gen.appBits; appBits > 0 {
		return uint((key & 0x3fffff) >> (randomBits - appBits))
	}
	return 0
}

// KeySeq returns the sequence number of the given key.
func (gen *Generator) KeySeq(key Key) uint {
	return uint((key & 0x3fffff)) & (1<<(randomBits-gen.appBits) - 1)
}

// MaxAppID returns the maximum application ID for `gen.Create(appID)`.
func (gen *Generator) MaxAppID() uint { return gen.appMax - 1 }
