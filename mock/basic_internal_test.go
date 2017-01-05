// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package mock

import (
	"bytes"
	"testing"
)

func TestReadWriteSeekableGrow(t *testing.T) {
	defer func() {
		err := recover()

		if err != bytes.ErrTooLarge {
			t.Errorf("Expected %v got %v", err, bytes.ErrTooLarge)
		}
	}()

	rws := &ReadWriteSeekable{}
	rws.grow(2147483647)
}
