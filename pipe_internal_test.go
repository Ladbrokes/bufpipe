// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package bufpipe

import (
	"io"
	"testing"

	"github.com/Ladbrokes/bufpipe/mock"
)

func TestClosedPipe(t *testing.T) {
	p := NewPipe(mock.NewReadWriteSeekable([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}))
	p.close()

	d := make([]byte, 10)
	if n, err := p.Read(d); err != io.ErrClosedPipe && n != 0 {
		t.Errorf("Reading a closed pipe. Expected [%v, %v] got [%v, %v]", 0, io.ErrClosedPipe, n, err)
	}

	if n, err := p.Write(d); err != io.ErrClosedPipe && n != 0 {
		t.Errorf("Writing a closed pipe. Expected [%v, %v] got [%v, %v]", 0, io.ErrClosedPipe, n, err)
	}
}
