// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

// Interfaces shared among go-csv and the Go standard library's encoding/csv.
// Can be used to easily use go-csv as a drop-in replacement for the latter.
package csv

// A helper interface for a general CSV reader. Conforms to encoding/csv Reader
// in the standard Go library as well as the Reader implemented by go-csv.
type Reader interface {
	// Read reads one record from r. The record is a slice of strings with each
	// string representing one field.
	Read() (record []string, err error)

	// ReadAll reads all the remaining records from r. Each record is a slice of
	// fields. A successful call returns err == nil, not err == EOF. Because
	// ReadAll is defined to read until EOF, it does not treat end of file as an
	// error to be reported.
	ReadAll() (records [][]string, err error)
}
