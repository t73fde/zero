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

package vocab

import (
	"bufio"
	"flag"
	"fmt"
	"iter"
	"math/rand/v2"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var wordlist = flag.String("wordlist", "",
	"file with one word per line; replaces the generated words in benchmarks")

// benchSizes are the current size of the index and the expected upper range.
var benchSizes = []int{49_000, 50_000, 195_000, 200_000}

// Sizes for the (long-running) searches.
var searchSizes = []int{50_000, 200_000}

// sink keeps results alive, so that calls cannot be optimized away.
var sink int

// --- corpus

var syllables = strings.Fields(`a ab al an ar ba be ber bo ca ch cor de den der di dis
	ed el en er es ex fa fi for ful ge ing ion ish ist ive ka ly ment mo na ness ni no
	or ous per pre pro re ri sa se sion sta ter tion tra un ver vi é ï`)

// genWord builds a word-like string. The length classes follow the measured
// distribution: about 25% up to 6 bytes, about 90% up to 14 bytes.
func genWord(rng *rand.Rand) string {
	var lo, hi int
	switch r := rng.IntN(100); {
	case r < 25:
		lo, hi = 2, 6
	case r < 90:
		lo, hi = 7, 14
	default:
		lo, hi = 15, 40
	}
	var sb strings.Builder
	for {
		sb.Reset()
		for sb.Len() < lo {
			sb.WriteString(syllables[rng.IntN(len(syllables))])
		}
		if sb.Len() <= hi {
			return sb.String()
		}
	}
}

// generateWords returns n distinct generated words.
func generateWords(n int) []string {
	rng := rand.New(rand.NewPCG(42, 43))
	seen := make(map[string]struct{}, n)
	words := make([]string, 0, n)
	for len(words) < n {
		w := genWord(rng)
		if _, dup := seen[w]; !dup {
			seen[w] = struct{}{}
			words = append(words, w)
		}
	}
	return words
}

// readWords returns all distinct words of the file at path.
func readWords(b *testing.B, path string, n int) []string {
	b.Helper()
	f, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { f.Close() }()
	seen := make(map[string]struct{})
	var words []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w := sc.Text()
		if _, dup := seen[w]; dup || len(w) > MaxWordLen {
			continue
		}
		seen[w] = struct{}{}
		words = append(words, w)
	}
	if err = sc.Err(); err != nil {
		b.Fatal(err)
	}
	if len(words) < n {
		b.Skipf("%s has %d distinct words, %d needed", path, len(words), n)
	}
	return words
}

var (
	corpusMu sync.Mutex
	corpora  = map[int][][]byte{}
)

// corpus returns n distinct words in random order; the result is cached.
func corpus(b *testing.B, n int) [][]byte {
	b.Helper()
	corpusMu.Lock()
	defer corpusMu.Unlock()
	if c, ok := corpora[n]; ok {
		return c
	}
	var words []string
	if *wordlist == "" {
		words = generateWords(n)
	} else {
		words = readWords(b, *wordlist, n)
	}
	rng := rand.New(rand.NewPCG(7, 8))
	rng.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
	c := make([][]byte, n)
	for i := range c {
		c[i] = []byte(words[i])
	}
	corpora[n] = c
	return c
}

// build returns a Vocabulary containing all words.
func build(words [][]byte) *Vocabulary {
	v := New(len(words))
	for _, w := range words {
		v.Add(w)
	}
	return v
}

// --- Add and Lookup

// BenchmarkAdd measures building a Vocabulary from scratch, with and without
// a sizeHint. Every iteration adds all words of the corpus.
func BenchmarkAdd(b *testing.B) {
	for _, n := range benchSizes {
		words := corpus(b, n)
		for _, hinted := range []bool{false, true} {
			b.Run(fmt.Sprintf("n=%d/hint=%v", n, hinted), func(b *testing.B) {
				b.ReportAllocs()
				hint := 0
				if hinted {
					hint = n
				}
				for range b.N {
					v := New(hint)
					for _, w := range words {
						v.Add(w)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*n), "ns/word")
			})
		}
	}
}

// BenchmarkAddKnown measures Add for words that are already present.
func BenchmarkAddKnown(b *testing.B) {
	for _, n := range benchSizes {
		words := corpus(b, n)
		v := build(words)
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			j := 0
			for range b.N {
				sink += int(v.Add(words[j]))
				if j++; j == n {
					j = 0
				}
			}
		})
	}
}

// BenchmarkLookup measures Lookup for present words (hit) and for absent words
// (miss). Keys are accessed in random order.
func BenchmarkLookup(b *testing.B) {
	for _, n := range benchSizes {
		words := corpus(b, n)
		v := build(words)
		misses := make([][]byte, n)
		for i, w := range words {
			misses[i] = append(w[:len(w):len(w)], '#') // '#' is not part of any word
		}
		for _, tc := range []struct {
			name string
			keys [][]byte
		}{{"hit", words}, {"miss", misses}} {
			b.Run(fmt.Sprintf("n=%d/%s", n, tc.name), func(b *testing.B) {
				b.ReportAllocs()
				j := 0
				for range b.N {
					sink += int(v.Lookup(tc.keys[j]))
					if j++; j == n {
						j = 0
					}
				}
			})
		}
	}
}

// BenchmarkLookupMap is the baseline for BenchmarkLookup: the same lookups in
// a map[string]WordID.
func BenchmarkLookupMap(b *testing.B) {
	for _, n := range benchSizes {
		words := corpus(b, n)
		m := make(map[string]WordID, n)
		misses := make([][]byte, n)
		for i, w := range words {
			m[string(w)] = WordID(i + 1)
			misses[i] = append(w[:len(w):len(w)], '#')
		}
		for _, tc := range []struct {
			kind string
			keys [][]byte
		}{{"hit", words}, {"miss", misses}} {
			b.Run(fmt.Sprintf("n=%d/%s", n, tc.kind), func(b *testing.B) {
				j := 0
				for range b.N {
					sink += int(m[string(tc.keys[j])]) // no allocation: compiler optimizes map[string(bytes)]
					if j++; j == n {
						j = 0
					}
				}
			})
		}
	}
}

var dummyV *Vocabulary
var dummyMap map[string]WordID
var dummyS string

// BenchmarkBuild compares build time of Vocabulary and map[string]WordID.
func BenchmarkBuild(b *testing.B) {
	for _, n := range []int{1_400, 10_300, 28_000, 76_100, 206_900} {
		words := corpus(b, n)
		b.Run(fmt.Sprintf("%d-vocab", n), func(b *testing.B) {
			for range b.N {
				v := New(len(words))
				for _, w := range words {
					v.Add(w)
				}
				dummyV = v
			}
		})
		b.Run(fmt.Sprintf("%d-map-o", n), func(b *testing.B) {
			for range b.N {
				m := make(map[string]WordID, len(words))
				for i, w := range words {
					m[string(w)] = WordID(i + 1)
				}
				dummyMap = m
			}
		})
		b.Run(fmt.Sprintf("%d-map-a", n), func(b *testing.B) {
			for range b.N {
				m := make(map[string]WordID, len(words))
				for i, w := range words {
					dummyS = string(w) // Often, there is a separate string conversion
					m[dummyS] = WordID(i + 1)
				}
				dummyMap = m
			}
		})
	}
}

// --- searches

type searchKind struct {
	name    string
	lengths []int // query lengths in bytes
	cut     func(w []byte, n int) []byte
	seq     func(*Vocabulary, []byte) iter.Seq[WordID]
	match   func(s, q string) bool // baseline on plain Go strings
}

var searchKinds = []searchKind{
	{name: "contains",
		lengths: []int{3, 5},
		cut:     func(w []byte, n int) []byte { k := (len(w) - n) / 2; return w[k : k+n] },
		seq:     (*Vocabulary).WordsContaining,
		match:   strings.Contains,
	},
	{name: "prefix",
		lengths: []int{3, cacheLen + 2}, // within and beyond the cache
		cut:     func(w []byte, n int) []byte { return w[:n] },
		seq:     (*Vocabulary).WordsWithPrefix,
		match:   strings.HasPrefix,
	},
	{name: "suffix",
		lengths: []int{3, 6},
		cut:     func(w []byte, n int) []byte { return w[len(w)-n:] },
		seq:     (*Vocabulary).WordsWithSuffix,
		match:   strings.HasSuffix,
	},
}

// queries returns 64 deterministic queries of n bytes, cut from random words.
func queries(words [][]byte, k searchKind, n int) (bs [][]byte, ss []string) {
	rng := rand.New(rand.NewPCG(11, 12))
	for len(bs) < 64 {
		w := words[rng.IntN(len(words))]
		if len(w) < n {
			continue
		}
		q := k.cut(w, n)
		bs = append(bs, q)
		ss = append(ss, string(q))
	}
	return bs, ss
}

// reportScan reports the cost per scanned word and the average number of hits.
func reportScan(b *testing.B, words, hits int) {
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*words), "ns/word")
	b.ReportMetric(float64(hits)/float64(b.N), "hits/op")
}

// BenchmarkSearch measures one complete scan per operation, for the Vocabulary
// and for a linear scan over a []string as baseline.
func BenchmarkSearch(b *testing.B) {
	for _, n := range searchSizes {
		words := corpus(b, n)
		v := build(words)
		strs := make([]string, n)
		for i, w := range words {
			strs[i] = string(w)
		}
		for _, k := range searchKinds {
			for _, length := range k.lengths {
				qb, qs := queries(words, k, length)
				name := fmt.Sprintf("n=%d/%s-%d", n, k.name, length)
				b.Run(name+"/vocab", func(b *testing.B) {
					hits := 0
					for i := range b.N {
						for range k.seq(v, qb[i%len(qb)]) {
							hits++
						}
					}
					reportScan(b, n, hits)
				})
				b.Run(name+"/strings", func(b *testing.B) {
					hits := 0
					for i := range b.N {
						q := qs[i%len(qs)]
						for _, s := range strs {
							if k.match(s, q) {
								hits++
							}
						}
					}
					reportScan(b, n, hits)
				})
			}
		}
	}
}

// --- memory

func heapAlloc() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// measureFootprint reports the heap growth caused by create, per word. The
// time per operation is meaningless here (it includes forced GCs).
func measureFootprint(b *testing.B, n int, create func() any) {
	var total int64
	for range b.N {
		before := heapAlloc()
		x := create()
		after := heapAlloc()
		total += int64(after) - int64(before)
		runtime.KeepAlive(x)
	}
	b.ReportMetric(float64(total)/float64(b.N*n), "B/word")
}

// BenchmarkFootprint compares the memory consumption of the Vocabulary with a
// map[string]WordID plus a []string for the reverse mapping (the model of a
// map-based store). Run with -benchtime=3x.
func BenchmarkFootprint(b *testing.B) {
	for _, n := range benchSizes {
		words := corpus(b, n)
		b.Run(fmt.Sprintf("n=%d/vocab", n), func(b *testing.B) {
			measureFootprint(b, n, func() any { return build(words) })
		})
		b.Run(fmt.Sprintf("n=%d/map", n), func(b *testing.B) {
			measureFootprint(b, n, func() any {
				m := make(map[string]WordID, n)
				list := make([]string, 0, n)
				for _, w := range words {
					s := string(w)
					list = append(list, s)
					m[s] = WordID(len(list))
				}
				return &struct {
					m    map[string]WordID
					list []string
				}{m, list}
			})
		})
	}
}

// BenchmarkLoadFactor measures building and Lookup for a hash table of fixed
// size (1<<16 positions) at different fill levels, up to the maximum load
// factor. It shows how the cost of Lookup and Add depends on the fill level
// and is meant to be rerun after changes to the table layout or the load
// factor constants.
func BenchmarkLoadFactor(b *testing.B) {
	const (
		tableSize = 1 << 16
		// Largest sizeHint that still yields a table of tableSize positions.
		maxHint = tableSize * loadNum / loadDen
	)
	if v := New(maxHint); len(v.hashes) != tableSize {
		b.Fatalf("table size %d, want %d", len(v.hashes), tableSize)
	}
	words := corpus(b, tableSize)
	for _, fill := range []float64{0.40, 0.55, 0.65, 0.75, float64(loadNum) / loadDen} {
		count := int(fill * tableSize)
		present := words[:count]
		absent := make([][]byte, count)
		for i, w := range present {
			absent[i] = append(w[:len(w):len(w)], '#') // '#' is not part of any word
		}
		v := New(maxHint)
		for _, w := range present {
			v.Add(w)
		}
		name := fmt.Sprintf("fill=%.3f", fill)
		for _, tc := range []struct {
			kind string
			keys [][]byte
		}{{"hit", present}, {"miss", absent}} {
			b.Run(name+"/"+tc.kind, func(b *testing.B) {
				j := 0
				for range b.N {
					sink += int(v.Lookup(tc.keys[j]))
					if j++; j == count {
						j = 0
					}
				}
			})
		}
		b.Run(name+"/build", func(b *testing.B) {
			for range b.N {
				vb := New(maxHint)
				for _, w := range present {
					vb.Add(w)
				}
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*count), "ns/word")
		})
	}
}
