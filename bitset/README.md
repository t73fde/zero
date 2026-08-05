# bitset

This package provides a compact set implementation for non-negative integers.

A `BitSet` stores each value as a single bit and is optimized for dense value ranges and fast membership tests.

Typical use cases include:

* character classes
* encoder escape tables
* bitmap containers

## Features

* Fast `Contains` checks
* Compact storage using machine words
* Efficient iteration over set values
* Cloning of independent copies
* Automatic storage growth
* Optional shrinking of unused storage

## Example

```go
package main

import (
	"fmt"

	"t73f.de/r/zero/bitset"
)

func main() {
	escapes := bitset.New(uint('*'), '[', ']', '\\')

	fmt.Println(escapes.Contains('*'))
	fmt.Println(escapes.Contains('a'))

	for value := range escapes.All() {
		fmt.Println(value)
	}
}
```

Output:

```
true
false
42
91
92
93
```

## API overview

### Creating sets

```go
bs := bitset.New(uint(1), 2, 100)
```

or from an iterator:

```go
bs := bitset.Collect(slices.Values([]uint{1, 2, 100}))
```

or by inserting values manually:

```go
var bs bitset.BitSet

bs.Insert(1)
bs.Insert(2)
bs.Insert(100)
```

### Modifying sets

```go
bs.Insert(42)
bs.Delete(42)
```

### Querying sets

```go
bs.Contains(42)

count := bs.Count()

lowest, ok := bs.Min()
highest, ok := bs.Max()
```

### Copying

`BitSet` values can be copied safely:

```go
clone := bs.Clone()
```

The clone has independent storage and can be modified without affecting the original.

A normal assignment only copies the `BitSet` value:

```go
alias := bs
```

Both values will refer to the same underlying storage.

### Iteration

Values are returned in ascending order:

```go
for value := range bs.All() {
	fmt.Println(value)
}
```

## Storage

`BitSet` is optimized for dense ranges of values.
For example, character sets or small integer domains are stored very efficiently.

For very sparse values with large gaps, a different data structure may be more appropriate.

## License

Licensed under the latest version of the EUPL (European Union Public License).

See the top-level `LICENSE.txt` file for details.
