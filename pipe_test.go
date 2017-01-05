// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package bufpipe_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Ladbrokes/bufpipe"
	"github.com/Ladbrokes/bufpipe/mock"
)

var errUnseekable = errors.New("Unseekable things")

func unseekableFunc(rws *mock.ReadWriteSeekable, offset int64, whence int) (int64, error) {
	return 0, errUnseekable
}

func TestNewPipe(t *testing.T) {
	p := bufpipe.NewPipe(&mock.ReadWriteSeekable{})
	if s, err := p.DataSize(); err != nil && s != 0 {
		t.Errorf("Empty object with empty data, expected [%v, %v] got [%v, %v]", nil, 0, err, s)
	}

	p = bufpipe.NewPipe(mock.NewReadWriteSeekable([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}))
	if s, err := p.DataSize(); err != nil && s != 10 {
		t.Errorf("Prepopulated object with data, expected [%v, %v] got [%v, %v]", nil, 10, err, s)
	}
}

func TestNewPipePanic(t *testing.T) {
	defer func() {
		err := recover()
		if err != errUnseekable {
			t.Errorf("Expected %v got %v", errUnseekable, err)
		}
	}()

	bufpipe.NewPipe(&mock.ReadWriteSeekable{SeekFunc: unseekableFunc})
}

func TestWriteReadPipe(t *testing.T) {
	p := bufpipe.NewPipe(&mock.ReadWriteSeekable{})

	if n, err := p.Write([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}); err != nil || n != 10 {
		t.Errorf("Issue writing, expected [%v, %v], got [%v, %v]", nil, 10, err, n)
	}

	if s, err := p.DataSize(); err != nil && s != 10 {
		t.Errorf("Issue querying data size after write, expected [%v, %v] got [%v, %v]", nil, 10, err, s)
	}

	d := make([]byte, 10)
	if n, err := p.Read(d); err != nil || n != 10 {
		t.Errorf("Issue reading, expected [%v, %v], got [%v, %v]", nil, 10, err, n)
	}

	if s, err := p.DataSize(); err != nil && s != 10 {
		t.Errorf("Issue querying data size after read, expected [%v, %v] got [%v, %v]", nil, 10, err, s)
	}
}

func TestBlockingReadPipe(t *testing.T) {
	p := bufpipe.NewPipe(&mock.ReadWriteSeekable{})

	worked := false

	// Don't exit the function until reading is done
	waitRead := make(chan struct{}, 1)
	// Don't try to write until reading is blocking
	waitWrite := make(chan struct{}, 1)

	go func() {
		defer close(waitRead)
		d := make([]byte, 10)

		// Ready to read, close channel to unblock main thread
		close(waitWrite)
		if n, err := p.Read(d); err != nil || n != 10 {
			t.Errorf("Issue reading, expected [%v, %v], got [%v, %v]", nil, 10, err, n)
		}
		worked = true
	}()

	// Block here until ready to read
	<-waitWrite

	// For good measure, take a nap
	time.Sleep(time.Millisecond)

	p.Write([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})

	// Wait for the read to finish
	<-waitRead

	if !worked {
		t.Error("Blocking read didn't work?")
	}
}

func TestReadSeekWriteReadPipe(t *testing.T) {
	p := bufpipe.NewPipe(mock.NewReadWriteSeekable([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}))

	d := make([]byte, 10)
	if n, err := p.Read(d); err != nil || n != 10 {
		t.Errorf("Issue reading, expected [%v, %v], got [%v, %v]", nil, 10, err, n)
	}

	if n, err := p.Seek(5, os.SEEK_SET); n != 5 || err != nil {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, nil, n, err)
	}

	p.Write([]byte{8, 7, 6})

	d = make([]byte, 10)

	n, err := p.Read(d)
	if err != nil || n != 8 {
		t.Errorf("Issue reading, expected [%v, %v], got [%v, %v]", nil, 8, err, n)
	}

	if expect := []byte{5, 6, 7, 8, 9, 8, 7, 6}; !bytes.Equal(d[:n], expect) {
		t.Errorf("Expected %v got %v", expect, d[:n])
	}

	for _, test := range [][]int{{22, os.SEEK_SET, 13}, {22, os.SEEK_END, 0}, {5, os.SEEK_CUR, 5}} {
		if n, err := p.Seek(int64(test[0]), test[1]); n != int64(test[2]) || err != nil {
			t.Errorf("Expected [%v, %v] got [%v, %v]", test[2], nil, n, err)
		}

	}
}

func TestLateSeekFailWritePipe(t *testing.T) {
	buf := mock.NewReadWriteSeekable([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	p := bufpipe.NewPipe(buf)
	buf.SeekFunc = unseekableFunc
	if n, err := p.Write([]byte{1}); n != 0 || err != errUnseekable {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, errUnseekable, n, err)
	}

	// Second call checks against werr
	if n, err := p.Write([]byte{1}); n != 0 || err != io.ErrClosedPipe {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, io.ErrClosedPipe, n, err)
	}
}

func TestLateSeekFailReadPipe(t *testing.T) {
	buf := mock.NewReadWriteSeekable([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	p := bufpipe.NewPipe(buf)
	buf.SeekFunc = unseekableFunc
	d := make([]byte, 1)
	if n, err := p.Read(d); n != 0 || err != errUnseekable {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, errUnseekable, n, err)
	}

	// Second call checks against rerr
	if n, err := p.Read(d); n != 0 || err != io.ErrClosedPipe {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, io.ErrClosedPipe, n, err)
	}
}

func TestSyncPipe(t *testing.T) {
	worked := false
	p := bufpipe.NewPipe(&mock.SyncReadWriteSeekable{ReadWriteSeekable: &mock.ReadWriteSeekable{}, SyncFunc: func() error {
		worked = true
		return nil
	}})

	if n, err := p.Write([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}); err != nil || n != 10 {
		t.Errorf("Issue writing, expected [%v, %v], got [%v, %v]", nil, 10, err, n)
	}

	if !worked {
		t.Errorf("Sync did not get called")
	}
}
