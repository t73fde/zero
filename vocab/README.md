# vocab

Package `vocab` stores a growing set of words in a compact form and assigns
each distinct word a stable, small integer ID (`WordID`). Words are found by
content, and all words that contain, start with, or end with a given byte
sequence can be enumerated.

It could be used for in-memory search indexes, where other data structures
should refer to a word by a 4-byte ID instead of holding their own string copy
(16 byte for the string descriptor) of the word.

```go
import "t73f.de/r/zero/vocab"
```

## When to use it

- Words are only added, never removed. To reclaim memory, build a new
  `Vocabulary`.
- Words are typically short and there are tens or hundreds of thousands of
  them, not billions.
- Contains, prefix and suffix searches may scan all words linearly.
- `Vocabulary` is expected to be used used by one goroutine at a time. It is
  **not** safe for concurrent use.

## Usage

```go
v := vocab.New(50_000) // expect about 50,000 distinct words

id := v.Add([]byte("House"))       // registers the word and returns its ID
same := v.Add([]byte("House"))     // known word: same ID
other := v.Lookup([]byte("Mouse")) // 0: not present, nothing is added

word := v.AppendBytes(nil, id) // ID -> bytes: "House"

// IDs of all words containing "ous":
for id := range v.WordsContaining([]byte("ous")) {
	fmt.Printf("%d: %s\n", id, v.AppendBytes(nil, id))
}
```

Limits can be lowered, for example to bound the memory consumption:

```go
v := vocab.New(0,
	vocab.WithMaxWords(500_000),
	vocab.WithMaxStoreLen(64<<20), // at most 64 MiB for words longer than 14 bytes
)
```

## Limits

| Limit | Default | Configurable | When exceeded |
|---|---|---|---|
| Word length (`MaxWordLen`) | 65535 bytes | no | `Add` panics; `Lookup` returns 0 |
| Number of words (`MaxWords`) | 2^32 - 1 | `WithMaxWords` | `Add` panics for a new word |
| Store for long words (`MaxStoreLen`) | 4 GiB - 1 | `WithMaxStoreLen` | `Add` panics for a new long word |

`Add` panics before it modifies the `Vocabulary`, so the state stays
consistent. Adding a known word never panics. Callers that want to avoid
panics can check `Len` and `StoreLen` beforehand.

## Concurrency

A `Vocabulary` has no internal locking and is not safe for concurrent use.
Serialize access, for example with a `sync.RWMutex` held around every call.
