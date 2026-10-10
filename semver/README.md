# semver

Package `semver` parses, validates, compares, and increments semantic version
strings according to [Semantic Versioning 2.0.0](https://semver.org).

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

## Install

```sh
go get t73f.de/r/zero/semver
```

## Usage

```go
v, ok := semver.Parse("1.4.2-rc.1+build.7")
if !ok {
	// invalid version string
}

fmt.Println(v)            // 1.4.2-rc.1+build.7
fmt.Println(v.IsValid())  // true

o := semver.MustParse("1.4.2") // panics on invalid input
fmt.Println(v.Compare(o))      // -1: a pre-release is lower than its release

v.IncMinor()
fmt.Println(v) // 1.5.0
```

## Notes

- A leading `v` (`v1.2.3`) is rejected.
- `MAJOR`, `MINOR`, and `PATCH` must fit into `int`; larger values are rejected.
