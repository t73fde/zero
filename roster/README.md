# roster

Package `roster` implements a set of non-negative integers as a sorted array.

Values are stored in strictly ascending order without duplicates. Membership
tests take O(log n); binary set operations merge both operands in a single
O(n+m) pass.

`roster` is the counterpart to `bitset`:

| | `bitset` | `roster` |
|---|---|---|
| Memory | proportional to the largest value | proportional to the number of values |
| Best for | dense value ranges | sparse sets, large values |

## Usage

```go
package main

import (
	"fmt"

	"example.com/yourmodule/roster"
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
}
```

## API overview

- Constructors: `New`, `Collect`
- Modification: `Insert`, `Delete`, `DeleteAll`, `Pop`
- Queries: `Contains`, `Count`, `IsEmpty`, `Min`, `Max`, `Equal`
- Iteration and output: `Values`, `String`, `WriteTo`
- Non-mutating set operations: `Union`, `Intersection`, `Difference`,
  `SymmetricDifference`, `Delta`
- Mutating set operations: `Or`, `And`, `AndNot`, `Xor`
- Memory management: `Clone`, `Grow`, `Clip`

## Complexity

| Operation | Time |
|---|---|
| `Contains` | O(log n) |
| `Insert`, `Delete` | O(n) |
| `Count`, `IsEmpty`, `Min`, `Max`, `Pop` | O(1) |
| `Union`, `Intersection`, `Difference`, `SymmetricDifference`, `Delta` | O(n+m) |
| `Or`, `And`, `AndNot`, `Xor` | O(n+m) |

To build a set from many values, use `New` or `Collect` instead of repeated
`Insert` calls.

## Copy semantics

Set operations never modify their operands, and their results never share
storage with them. A `Roster` is a value type holding a slice: a plain
assignment (`b := a`) copies only the slice header, and both values then
share the same storage. An independent copy is obtained only via `Clone`.