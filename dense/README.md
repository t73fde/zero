# dense

Package `dense` provides a bijection between comparable keys and dense
unsigned integer codes.

Codes are assigned in order of first use, starting at 0, without gaps. This
makes them suitable as elements of bit sets or sorted-array sets (see
`bitset` and `roster`), as indices into slices, or as node numbers of a
graph.

## Usage

```go
import "t73f.de/r/zero/dense"

var d dense.Dense[string, uint32] // the zero value is ready to use

a := d.Encode("alpha")  // 0, assigned
b := d.Encode("beta")   // 1, assigned
_ = d.Encode("alpha")   // 0, already known

code, ok := d.Find("gamma") // 0, false: unknown keys are not assigned
key, ok := d.Decode(b)      // "beta", true
n := d.Count()              // 2
```


## Properties and limits

- The mapping is a bijection between the keys seen so far and `0 … Count()-1`.
- Codes follow the order of insertion, not the order of the keys.
- `Encode` panics if the range of `V` is exhausted, e.g. the 257th key for
  `uint8`. Known keys stay accessible.
- Not safe for concurrent use. Do not copy a `Dense` after first use.