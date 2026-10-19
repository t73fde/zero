# dense

Package `dense` provides a bijection between comparable keys and dense
unsigned integer values.

Values are assigned in order of first use, starting at 0. Freed values are
reused, so the values in use stay close to `0 ... Count()-1`. This makes them
suitable as elements of bit sets or sorted-array sets (see `bitset` and
`roster`), as indices into slices, or as node numbers of a graph.

## Usage

```go
import "t73f.de/r/zero/dense"

var d dense.Dense[string, uint32] // the zero value is ready to use

d.Encode("alpha")      // 0, assigned
b := d.Encode("beta")  // 1, assigned
d.Encode("alpha")      // 0, already known

val, ok := d.Find("gamma") // 0, false: unknown keys are not assigned
key, ok := d.Decode(b)     // "beta", true
n := d.Count()             // 2

val, ok = d.RemoveKey("beta") // 1, true: value 1 is freed for reuse
d.Encode("gamma")             // 1, reuses the freed value
```

## Properties and limits

- The mapping is a bijection between the keys currently present and the
  values in use; all values in use are less than `Limit()`.
- Values follow neither the order of the keys nor, after a removal, the
  order of insertion.
- `Encode` panics if a new key needs a value and the range of `V` is
  exhausted, e.g. a 257th key for `uint8`. Known keys stay accessible.
- Not safe for concurrent use. Do not copy a `Dense` after first use.

See the package documentation for the handling of removed values and for
the restrictions on keys (e.g. NaN).