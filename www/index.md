# App - Utilities for Go applications

This is a collection of utility functions to build applications in
[Go](https://go.dev/).

* [iter](/dir?ci=tip&name=iter): additional functions to combine iterators.
* [set](/dir?ci=tip&name=set): a simple set type.

## Use instructions

If you want to import this library into your own [Go](https://go.dev/)
software, you must execute a `go get` command. Since Go treats non-standard
software and non-standard platforms quite badly, you must use some non-standard
commands.

First, you must install the version control system
[Fossil](https://fossil-scm.org), which is a superior solution compared to Git,
in too many use cases. It is just a single executable, nothing more. Make sure,
it is in your search path for commands.

How you can execute the following Go command to retrieve a given version of
this library:

    GOVCS=t73f.de:fossil go get t73f.de/r/app@HASH

where `HASH` is the hash value of the commit you want to use.

Go currently seems not to support software versions when the software is
managed by Fossil. This explains the need for the hash value. However, this
methods works.
