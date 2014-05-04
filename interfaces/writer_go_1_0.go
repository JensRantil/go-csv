// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

// +build !go1.1

package interfaces

// A helper interface for a general CSV writer. Conforms to encoding/csv Writer
// in the standard go library as well as the Writer implemented by this
// package.
type Writer interface {
	// Flush writes any buffered data to the underlying io.Writer.
	// To check if an error occurred during the Flush, call Error.
	Flush()

	// Writer writes a single CSV record to w along with any necessary quoting.
	// A record is a slice of strings with each string being one field.
	Write(record []string) error

	// WriteAll writes multiple CSV records to w using Write and then calls Flush.
	WriteAll(records [][]string) error
}
