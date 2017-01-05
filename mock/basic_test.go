// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package mock_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/Ladbrokes/bufpipe/mock"
)

const leetLen = 1337

func testrw(t *testing.T, p []byte, f func([]byte) (n int, err error)) {
	n, err := f(p)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(p) {
		t.Errorf("Expected %d got %d", len(p), n)
	}
}

func TestReadWriteSeekable(t *testing.T) {
	dataSet := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	dsl := len(dataSet)
	p := make([]byte, dsl)
	rws := mock.NewReadWriteSeekable(dataSet)

	if !bytes.Equal(dataSet, rws.Bytes()) {
		t.Error("Expected Bytes to return the buffer...")
	}

	testrw(t, dataSet, rws.Write)
	testrw(t, p, rws.Read)

	if !bytes.Equal(dataSet, p) {
		t.Errorf("Expected %v got %v", dataSet, p)
	}

	for _, test := range [][]int64{{-1, 0}, {11, 10}, {0, 0}, {10, 10}, {5, 5}} {
		s, err := rws.Seek(test[0], os.SEEK_SET)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if s != test[1] {
			t.Errorf("Expected %v got %v", test[1], s)
		}
	}

	for _, test := range [][]int64{{0, 5}, {-1, 4}, {2, 6}, {50, 10}, {-50, 0}} {
		s, err := rws.Seek(test[0], os.SEEK_CUR)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if s != test[1] {
			t.Errorf("Expected %v got %v", test[1], s)
		}
	}

	for _, test := range [][]int64{{0, 10}, {-1, 10}, {2, 8}, {50, 0}, {-50, 10}, {5, 5}} {
		s, err := rws.Seek(test[0], os.SEEK_END)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if s != test[1] {
			t.Errorf("Expected %v got %v", test[1], s)
		}
	}

	testrw(t, []byte{0}, rws.Write)
	rws.Seek(0, os.SEEK_SET)
	testrw(t, p, rws.Read)

	if expect := []byte{0, 1, 2, 3, 4, 0, 6, 7, 8, 9}; !bytes.Equal(expect, p) {
		t.Errorf("Expected %v got %v", expect, p)
	}

	rws.Seek(5, os.SEEK_SET)
	testrw(t, []byte{0, 1, 2, 3, 4, 5, 6}, rws.Write)
	rws.Seek(4, os.SEEK_SET)
	p = make([]byte, 8)
	testrw(t, p, rws.Read)

	if expect := []byte{4, 0, 1, 2, 3, 4, 5, 6}; !bytes.Equal(expect, p) {
		t.Errorf("Expected %v got %v", expect, p)
	}

	rws.Seek(0, os.SEEK_SET)
	p = make([]byte, 12)
	testrw(t, p, rws.Read)
	if expect := []byte{0, 1, 2, 3, 4, 0, 1, 2, 3, 4, 5, 6}; !bytes.Equal(expect, p) {
		t.Errorf("Expected %v got %v", expect, p)
	}

	bigger := make([]byte, 70)
	rws.Seek(0, os.SEEK_END)
	testrw(t, bigger, rws.Write)
	if l := rws.Len(); l != 82 {
		t.Errorf("Expected %v got %v", 82, l)
	}

	rws.Reset()
	s, err := rws.Seek(0, os.SEEK_END)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if s != 0 {
		t.Errorf("Expected %v got %v", 0, s)
	}

	n, err := rws.Read(p)
	if err != io.EOF {
		t.Errorf("Expected %v got %v", io.EOF, err)
	}
	if n != 0 {
		t.Errorf("Expected %v got %v", 0, n)
	}
}

func TestReadWriteSeekableOverrides(t *testing.T) {
	worked := false

	errReadTest := errors.New("read test")
	errWriteTest := errors.New("write test")
	errSeekTest := errors.New("seek test")

	p := &mock.ReadWriteSeekable{
		ReadFunc: func(rws *mock.ReadWriteSeekable, b []byte) (n int, err error) {
			worked = true
			return leetLen, errReadTest
		},
		WriteFunc: func(rws *mock.ReadWriteSeekable, b []byte) (n int, err error) {
			worked = true
			return leetLen, errWriteTest
		},
		SeekFunc: func(rws *mock.ReadWriteSeekable, offset int64, whence int) (int64, error) {
			worked = true
			return leetLen, errSeekTest
		},
		LenFunc: func(rws *mock.ReadWriteSeekable) int {
			worked = true
			return leetLen
		},
	}

	b := make([]byte, 1)

	worked = false
	if n, err := p.Read(b); n != leetLen || err != errReadTest || !worked {
		t.Errorf("Read test expected [%v, %v, %v] got [%v, %v, %v]", leetLen, errReadTest, true, n, err, worked)
	}

	worked = false
	if n, err := p.Write(b); n != leetLen || err != errWriteTest || !worked {
		t.Errorf("Write test expected [%v, %v, %v] got [%v, %v, %v]", leetLen, errWriteTest, true, n, err, worked)
	}

	worked = false
	if n, err := p.Seek(leetLen, os.SEEK_SET); n != leetLen || err != errSeekTest || !worked {
		t.Errorf("Seek test expected [%v, %v, %v] got [%v, %v, %v]", leetLen, errSeekTest, true, n, err, worked)
	}

	worked = false
	if n := p.Len(); n != leetLen || !worked {
		t.Errorf("Len test expected [%v, %v] got [%v, %v]", leetLen, true, n, worked)
	}

}
