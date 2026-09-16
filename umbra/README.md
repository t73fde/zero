# umbra

Package `umbra` provides a compact string representation (aka "german strings"),
aimed at in-memory search indexes that store large numbers of short words and
word fragments.

## Background

Instead of a plain pointer+length string, each value is packed into a fixed
16-byte struct: 2 bytes length, 14 bytes payload. Strings up to 14 bytes are
stored entirely inline, with no separate allocation and no indirection. Longer
strings store a cached prefix inline plus an offset into a shared arena, so
most comparisons never touch the arena at all.

The package is named after the
[_Umbra Paper_](https://db.in.tum.de/~freitag/papers/p29-neumann-cidr20.pdf),
which introduced the idea of this type of string representation.

## Constraints

- Individual strings are limited to 65,535 bytes (2**16).
- Arenas are limited to 4,294,967,296 bytes (2**32).
- Designed for workloads with many short, frequently repeated fragments (e.g.
  tokenized words) rather than general-purpose text storage.
- Not safe to use a string with any arena other than the one it was created
  from.
