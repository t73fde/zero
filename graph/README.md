# graph

Package `graph` implements a directed graph of orderable values
(`cmp.Ordered`) on top of `t73f.de/r/zero/set`.

```
import "t73f.de/r/zero/graph"
```

## Data model

```go
type Digraph[T cmp.Ordered] map[T]set.Set[T]
```

A digraph maps each vertex to the set of its direct successors.

- Every vertex is a key of the map, including vertices without outgoing edges.
- The zero value (`nil`) is an empty digraph and is usable with all methods.
  Methods that may add vertices return the (possibly newly allocated)
  digraph and must be used as `dg = dg.AddEdge(a, b)`.
- Methods returning a digraph return `nil` for an empty result.
- Methods returning a set always return a non-nil set.
- Only forward edges are stored. Predecessor queries are answered via
  `Reverse()`, which costs O(V+E) per call.
- A `Digraph` is not safe for concurrent modification.

```go
type Edge[T cmp.Ordered] struct{ From, To T }
type EdgeSlice[T cmp.Ordered] []Edge[T]
```

## Example

```go
var dg graph.Digraph[string]
dg = dg.AddEdges(graph.EdgeSlice[string]{
	{From: "a", To: "b"},
	{From: "a", To: "c"},
	{From: "b", To: "d"},
	{From: "c", To: "d"},
})

if _, ok := dg.IsDAG(); ok {
	fmt.Println(dg.SortReverse())                       // [d c b a]
}
fmt.Println(dg.ReachableVertices("b"))                  // set {d}
fmt.Println(dg.TransitiveClosure("c").Edges().Sort())   // [{c d}]
fmt.Println(dg.Reverse().Edges().Sort())                // [{b a} {c a} {d b} {d c}]
```
