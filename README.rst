CSV
===
A Go CSV_ implementation inspired by `Python's CSV module`_. Supports custom
CSV formats. Currently only writing CSV files is supported.

.. _CSV: https://en.wikipedia.org/wiki/Comma-separated_values
.. _Python's CSV module: https://docs.python.org/2/library/csv.html

Examples
--------
Here's a basic example::

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

To modify CSV dialect, have a look at ``csv.Dialect``. It supports changing:

* separator/delimiter.

* quoting modes:
  
  * Always quote.
   
  * Never quote.
   
  * Quote when needed (minimal quoting).

  * Quote all non-numerical fields.

* line terminator.

* how quote character escaping should be done - using double escape, or using a
  custom escape character.

Have a look at documentation_ or ``csv_test.go`` for example on how to use
these. All values above have sane defaults (that makes the module behave the
same as the ``csv`` module in the Go standard library).

.. _documentation: http://godoc.org/github.com/JensRantil/go-csv

Documentation
-------------
Package documentation can be found here_.

.. _here: http://godoc.org/github.com/JensRantil/go-csv

Why was this developed?
-----------------------
I needed it for mysqlcsvdump_ to support variations of CSV output. The ``csv``
module in the Go (1.2) standard library was inadequate as it it does not
support any CSV dialect modifications except changing separator and partially
line termination.

.. _mysqlcsvdump: https://github.com/JensRantil/mysqlcsvdump

Who developed this?
-------------------
I'm Jens Rantil. Have a look at `my blog`_ for more info on what I'm working
on.

.. _my blog: http://jensrantil.github.io/pages/about-jens.html
