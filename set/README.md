# Package set

Package `set` implements a set of values of a comparable type as a hash map.

```go
import "t73f.de/r/zero/set"
```

## Overview

```go
type Set[V comparable] struct { /* unexported */ }
```

A `Set` stores each value at most once. Values that are equal under `==` are
the same element. The zero value is an empty set that is ready to use.

`Set` offers the same API as package `roster`, with two differences that
follow from the hash map: there is no order, and therefore no `Min` and `Max`.


## When to use which package

|                     | `bitset`                    | `roster`                       | `set`                      |
|---------------------|-----------------------------|--------------------------------|----------------------------|
| Element type        | non-negative integers       | `cmp.Ordered`                  | `comparable`               |
| Memory              | proportional to max value   | proportional to number of values | proportional to number of values, with higher constant |
| `Contains`          | O(1)                        | O(log n)                       | O(1) expected              |
| `Insert`, `Delete`  | O(1)                        | O(n)                           | O(1) expected              |
| Set operations      | O(max/64)                   | O(n+m), single merge pass      | O(n+m) expected            |
| Iteration order     | ascending                   | ascending                      | unspecified                |

Use `set` if the element type has no order, or if elements are inserted and
deleted one by one at a large scale. Use `roster` if sets are built once and
then mostly read, combined, or iterated in order. Use `bitset` for dense
ranges of small integers.

## Semantics

- **Copying.** Assigning a `Set` or passing it by value copies only the map
  reference. After modifying one copy, the state of the other copy is
  unspecified. An independent copy is obtained only by `Clone`.
- **Receivers.** Methods that modify the set have pointer receivers, because
  the map of a zero value is allocated on first use. A `Set` stored in a map
  or slice element must be modified via a variable and written back, or
  `*Set` values must be stored.
- **Order.** `Values`, `String`, `WriteTo`, and `Pop` use an unspecified
  order. To iterate in order, collect the values and sort them with
  `slices.Sorted(s.Values())`.
- **Floating-point values.** `0.0` and `-0.0` are one element. `NaN` is never
  equal to anything, so every `Insert(NaN)` adds a new element that cannot be
  found or deleted. Do not use `Set` for floating-point values that may be
  `NaN`.
- **Interface types.** If `V` is an interface type, inserting a value whose
  dynamic type is not comparable panics at run time.
- **Memory.** A Go map does not report its capacity and does not release
  memory when values are deleted. `Grow` and `Shrink` therefore always
  allocate a new map and copy the values. `DeleteAll` keeps the allocated
  storage.
- **`Pop`.** `Pop` removes an arbitrary value. On large, sparsely filled maps
  a single call may take longer than O(1).
