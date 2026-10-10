# zero/iter

Helpers for working with Go iterators (`iter.Seq`, `iter.Seq2`).

```go
import zeroiter "t73f.de/r/zero/iter"
```

The package name equals the standard library package `iter`; use an import
alias ("zeroiter" preferred) if both are needed.

All iterators are lazy and stop as soon as the consumer stops.

## Example

```go
squares := zeroiter.MapSeq(
	zeroiter.TakeSeq(5, zeroiter.CountSeq()),
	func(i int) int { return i * i },
)
fmt.Println(slices.Collect(squares)) // [0 1 4 9 16]
```
