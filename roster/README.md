# roster

Package `roster` implements a set of ordered values (`cmp.Ordered`) as a
sorted array.

Values are stored in strictly ascending order without duplicates. Membership
tests take O(log n) comparisons. Binary set operations merge both operands in
a single pass of O(n+m) comparisons.

`roster` is the counterpart to `bitset` if the element type is restricted to
non-negative integer values:

| | `bitset` | `roster` |
|---|---|---|
| Element types | non-negative integers | any `cmp.Ordered` type, e.g. integers, strings |
| Memory | proportional to the largest value | proportional to the number of values |
| Best for | dense value ranges | sparse sets, large values, non-integer values |

## Usage

```go
package main

import (
	"fmt"

	"t73f.de/r/zero/roster"
)

func main() {
	oldSet := roster.New[uint](1, 3, 5, 7)
	newSet := roster.New[uint](3, 4, 5, 9)

	removed, added := oldSet.Delta(newSet)
	fmt.Println(removed) // 1 7
	fmt.Println(added)   // 4 9

	var s roster.Roster[uint32] // zero value is an empty set
	s.Insert(42)
	fmt.Println(s.Contains(42)) // true

	for v := range s.Values() { // ascending order
		fmt.Println(v)
	}

	words := roster.New("pear", "apple", "fig")
	fmt.Println(words) // apple fig pear
}
```

## Ordering and equality

Values are ordered by `cmp.Compare`; values that compare equal are the same
element. For floating-point types, `0.0` and `-0.0` are one element, and
`NaN` equals `NaN` and sorts before all other values. Strings are compared
bytewise, so a comparison costs time proportional to the length of the common
prefix.

## Copy semantics

Set operations never modify their operands, and their results never share
storage with them. A `Roster` is a value type holding a slice: a plain
assignment (`b := a`) copies only the slice header, and both values then
share the same storage. After modifying one of them, the state of the other
is unspecified. An independent copy is obtained only via `Clone`.