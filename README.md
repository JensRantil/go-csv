CSV
===
[![Build Status](https://secure.travis-ci.org/JensRantil/go-csv.png?branch=master)](http://travis-ci.org/JensRantil/go-csv) [![Go Report Card](https://goreportcard.com/badge/github.com/JensRantil/go-csv)](https://goreportcard.com/report/github.com/JensRantil/go-csv) [![GoDoc](https://godoc.org/github.com/JensRantil/go-csv?status.svg)](https://godoc.org/github.com/JensRantil/go-csv)

A Go [CSV](https://en.wikipedia.org/wiki/Comma-separated_values) implementation
inspired by [Python's CSV module](https://docs.python.org/2/library/csv.html).
It supports various CSV dialects (see below) and is fully backward compatible
with the [`encoding/csv`](http://golang.org/pkg/encoding/csv/) package in the
Go standard library.

Examples
--------

Writing
~~~~~~~
Here's a basic writing example:

```go
f, err := os.Create("output.csv")
checkError(err)
defer func() {
  err := f.Close()
  checkError(err)
}
w := NewWriter(f)
w.Write([]string{
  "a",
  "b",
  "c",
})
w.Flush()
// output.csv will now contains the line "a b c" with a trailing newline.
```

Reading
~~~~~~~
Here's a basic reading example:

```go
f, err := os.Open('myfile.csv')
checkError(err)
defer func() {
  err := f.Close()
  checkError(err)
}

r := NewReader(f)
for {
  fields, err := r.Read()
  if err == io.EOF {
    break
  }
  checkOtherErrors(err)
  handleFields(fields)
}
```

CSV dialects
------------
To modify CSV dialect, have a look at `csv.Dialect`,
`csv.NewDialectWriter(...)` and `csv.NewDialectReader(...)`. It supports
changing:

* separator/delimiter.
* quoting modes:
  * Always quote.
  * Never quote.
  * Quote when needed (minimal quoting).
  * Quote all non-numerical fields.
* line terminator.
* how quote character escaping should be done - using double escape, or using a
  custom escape character.

Have a look at [the
documentation](http://godoc.org/github.com/JensRantil/go-csv) in `csv_test.go`
for example on how to use these. All values above have sane defaults (that
makes the module behave the same as the `csv` module in the Go standard library).

Documentation
-------------
Package documentation can be found
[here](http://godoc.org/github.com/JensRantil/go-csv).

Why was this developed?
-----------------------
I needed it for [mysqlcsvdump](https://github.com/JensRantil/mysqlcsvdump) to
support variations of CSV output. The `csv` module in the Go (1.2) standard
library was inadequate as it it does not support any CSV dialect modifications
except changing separator and partially line termination.

Who developed this?
-------------------
I'm Jens Rantil. Have a look at [my
blog](http://jensrantil.github.io/pages/about-jens.html) for more info on what
I'm working on.
