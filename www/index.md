# Zero - Utilities for Go applications

This is a collection of utility functions for building applications in
[Go](https://go.dev/).

* [graph](/dir?ci=tip&name=graph): A simple directed graph.
* [iter](/dir?ci=tip&name=iter): Additional functions for combining iterators.
* [set](/dir?ci=tip&name=set): A simple set type.
* [snow](/dir?ci=tip&name=snow): A key generator.

The name _Zero_ refers to "0-components", introduced in the book
_Moderne Software-Architektur: Umsichtig planen, robust bauen mit Quasar_ by
Johannes Siedersleben (2004). 0-components are independent of both application
and technology, much like blood type O, which can be universally transfused.

## Usage instructions

To import this library into your own [Go](https://go.dev/) software, you need
to run the `go get` command. Since Go does not handle non-standard software and
platforms well, some additional steps are required.

First, install the version control system [Fossil](https://fossil-scm.org),
which is a superior alternative to Git in many use cases. Fossil is just a
single executable, nothing more. Make sure it is included in your system's
command search path.

Then, run the following Go command to retrieve a specific version of
this library:

    GOVCS=t73f.de:fossil go get t73f.de/r/zero@HASH

Here, `HASH` represents the commit hash of the version you want to use.

Go currently does not seem to support software versioning for projects managed
by Fossil. This is why the hash value is required. However, this method works
reliably.
