# slices

Package `slices` provides functions for slices of any type that are missing in
the standard package [`slices`](https://pkg.go.dev/slices).

Import path: `t73f.de/r/zero/slices`. Since the name equals the standard
package, an import alias is needed when both are used, e.g. `zeroslices`.

## Functions

### Shrink

```go
func Shrink[E any](s []E) []E
```

Releases unused storage. A Go slice keeps its backing array, even if values
were deleted. `Shrink` copies the values to a new slice of exactly the needed
size, so the old backing array can be garbage collected, provided no other
reference to it exists.

`Shrink` reallocates and is therefore relatively expensive. Use it for
long-living slices that were built from larger intermediate results or that
lost many values.

```go
s := make([]int, 0, 1_000_000)
s = append(s, 1, 2, 3)
s = zslices.Shrink(s) // len 3, cap 3, old array is collectable
```