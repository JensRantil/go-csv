// Copyright 2014 Jens Rantil. All rights reserved.  Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package csv

import (
	"testing"
)

func TestIsNumeric(t *testing.T) {
	t.Parallel()

	notNumeric := []string{
		"",
		" ",
		"a",
		"1a",
		"a1",
	}
	numeric := []string{
		"1",
		"11",
		"123456789",
	}
	for _, item := range numeric {
		if !isNumeric(item) {
			t.Error("Should be numeric:", item)
		}
	}
	for _, item := range notNumeric {
		if isNumeric(item) {
			t.Error("Should not be numeric:", item)
		}
	}
}
