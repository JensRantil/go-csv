// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package interfaces

import (
	"bytes"
	oldcsv "encoding/csv"
	thiscsv "github.com/JensRantil/go-csv"
	"testing"
)

func TestWriterInterface(t *testing.T) {
	t.Parallel()

	var iface Writer
	iface = thiscsv.NewWriter(new(bytes.Buffer))
	iface = thiscsv.NewDialectWriter(new(bytes.Buffer), thiscsv.Dialect{})
	iface = oldcsv.NewWriter(new(bytes.Buffer))

	// To get rid of compile-time warning that this variable is not used.
	iface.Flush()
}
