# zero/snow

Package `snow` generates unique, time-ordered 64-bit keys (Snowflake/TSID
style). A key can be used as an element of a URI, as a primary key in a
database, or wherever a compact, sortable identifier is needed.

```
go get t73f.de/r/zero/snow
```

## Key layout

A `Key` is a `uint64`:

```
 63                                        22 21                      0
+--------------------------------------------+-------------------------+
| timestamp (42 bit)                         | app ID | sequence       |
+--------------------------------------------+-------------------------+
                                               appBits  22 - appBits
```

* Timestamp: milliseconds since 2024-06-01 00:00:00 UTC. 42 bits last until
  the year 2163.
* Application ID: 0 to 20 bits (`MaxAppBits`), chosen when creating a
  `Generator`. Intended for application-defined data, e.g. a table number or a
  node number.
* Sequence: the remaining bits, at least 2. Distinguishes keys generated
  within the same millisecond.

Because the timestamp occupies the most significant bits, keys sort by
creation time. `Invalid` (value 0) is never generated.

Number of keys per millisecond and generator: `1 << (22 - appBits)`, e.g.
4,194,304 for `appBits = 0` and 4 for `appBits = 20`. If the sequence space of
the current millisecond is exhausted, `Create` waits for the next millisecond.

## Usage

```go
package main

import (
	"fmt"

	"t73f.de/r/zero/snow"
)

func main() {
	gen := snow.New(4) // 4 bits for application data: IDs 0..15

	key := gen.Create(3)
	fmt.Println(key)                 // e.g. 1F2G3H4J5K6M7
	fmt.Println(key.Time())          // creation time
	fmt.Println(gen.AppID(key))      // 3
	fmt.Println(gen.KeySeq(key))     // sequence number within the millisecond
	fmt.Println(key.Format(4, "-"))  // 1-F2G3-H4J5-K6M7

	parsed, err := snow.Parse(key.Format(4, "-"))
	if err != nil {
		panic(err)
	}
	fmt.Println(parsed == key) // true
}
```

## String representation

Keys are encoded as base-32 using Crockford's alphabet
(`0123456789ABCDEFGHJKMNPQRSTVWXYZ`).

* `String` returns the shortest form without leading zeroes, at most 13
  characters. `Invalid` is `"0"`.
* `Format(groupSize, sep)` returns all 13 digits, grouped by `sep`, e.g.
  `Invalid.Format(4, "-") == "0-0000-0000-0000"`.
* `Parse` accepts upper and lower case letters, the aliases `I`, `L` (as `1`)
  and `O` (as `0`), and single `-` between digits. It returns an error for an
  empty string, for other characters, and for values that do not fit into 64
  bits (the first of 13 digits must be at most `F`).

## Properties and limits

* A `Generator` is safe for concurrent use. Keys are unique per generator.
  Different generators (e.g. on different nodes) must use different
  application IDs to avoid collisions.
* The clock must not be set before 2024-06-01 UTC. `Create` panics if the
  timestamp is outside the range of 42 bits.
* If the system clock moves backwards, `Create` continues with the last used
  timestamp, so keys remain unique and monotonic.
* Keys are not random and not secret. They disclose the creation time and the
  application ID.