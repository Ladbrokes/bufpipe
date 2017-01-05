// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package bufpipe_test

import (
	"bufio"
	"bytes"
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/Ladbrokes/bufpipe"
	"github.com/Ladbrokes/bufpipe/mock"
)

func unwriteableFunc(rws *mock.ReadWriteSeekable, b []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestNewLineIndexedPipe(t *testing.T) {
	p := bufpipe.NewLineIndexedPipe(&mock.ReadWriteSeekable{}, &mock.ReadWriteSeekable{})
	if s, err := p.DataSize(); err != nil && s != 0 {
		t.Errorf("Empty object with empty data, expected [%v, %v] got [%v, %v]", nil, 0, err, s)
	}
}

func TestWriteReadIndexedPipe(t *testing.T) {
	data := &mock.ReadWriteSeekable{}
	index := &mock.ReadWriteSeekable{}
	p := bufpipe.NewLineIndexedPipe(data, index)

	testBytes := []byte("\nHello World\n\n")
	expectIndex := []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 13, 0, 0, 0, 0, 0, 0, 0}

	if n, err := p.Write(testBytes); err != nil || n != 14 {
		t.Errorf("Expected [%v, %v], got [%v, %v]", nil, 14, err, n)
	}

	if n, err := p.DataSize(); err != nil || n != 14 {
		t.Errorf("Expected [%v, %v], got [%v, %v]", nil, 14, err, n)
	}

	// 8 bytes * 3 lines
	if n, err := p.IndexSize(); err != nil || n != 24 {
		t.Errorf("Expected [%v, %v], got [%v, %v]", nil, 24, err, n)
	}

	if n, err := p.CountLines(); err != nil || n != 3 {
		t.Errorf("Expected [%v, %v], got [%v, %v]", nil, 3, err, n)
	}

	if !bytes.Equal(testBytes, data.Bytes()) {
		t.Errorf("Expected %v got %v", testBytes, data)
	}

	if !bytes.Equal(expectIndex, index.Bytes()) {
		t.Errorf("Expected %v got %v", expectIndex, index)
	}

	if err := p.SeekLine(1); err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	scanner := bufio.NewScanner(p)

	scanner.Scan()
	if got := scanner.Text(); got != "Hello World" {
		t.Errorf("Expected Hello World got %v", got)
	}

	scanner.Scan()
	if got := scanner.Text(); got != "" {
		t.Errorf("Expected empty string got %v", got)
	}

	index.SeekFunc = unseekableFunc
	if err := p.SeekLine(1); err != errUnseekable {
		t.Errorf("Expected %v got %v", errUnseekable, err)
	}

	index.SeekFunc = nil
	if n, err := p.Read(testBytes); n != 0 || err != io.ErrClosedPipe {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, io.ErrClosedPipe, n, err)
	}

	if n, err := p.Write(testBytes); n != 0 || err != io.ErrClosedPipe {
		t.Logf("Expected [%v, %v] got [%v, %v]", 0, io.ErrClosedPipe, n, err)
	}

}

func TestLateDataSeekFailWriteIndexedPipe(t *testing.T) {
	testBytes := []byte("\nHello World\n\n")
	data := &mock.ReadWriteSeekable{}
	index := &mock.ReadWriteSeekable{}
	p := bufpipe.NewLineIndexedPipe(data, index)
	data.SeekFunc = unseekableFunc

	if n, err := p.Write(testBytes); n != 0 || err != errUnseekable {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, errUnseekable, n, err)
	}

	data.SeekFunc = nil
	if n, err := p.Read(testBytes); n != 0 || err != errUnseekable {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, errUnseekable, n, err)
	}
	if n, err := p.Write(testBytes); n != 0 || err != errUnseekable {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, errUnseekable, n, err)
	}
}

func TestLateIndexSeekFailWriteIndexedPipe(t *testing.T) {
	testBytes := []byte("\nHello World\n\n")
	data := &mock.ReadWriteSeekable{}
	index := &mock.ReadWriteSeekable{}
	p := bufpipe.NewLineIndexedPipe(data, index)
	index.SeekFunc = unseekableFunc

	if n, err := p.Write(testBytes); n != 0 || err != errUnseekable {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, errUnseekable, n, err)
	}

	data.SeekFunc = nil
	if n, err := p.Write(testBytes); n != 0 || err != io.ErrClosedPipe {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, io.ErrClosedPipe, n, err)
	}
}

func TestIndexWriteFailureIndexedPipe(t *testing.T) {
	testBytes := []byte("\nHello World\n\n")
	data := &mock.ReadWriteSeekable{}
	index := &mock.ReadWriteSeekable{}
	p := bufpipe.NewLineIndexedPipe(data, index)
	index.WriteFunc = unwriteableFunc

	if n, err := p.Write(testBytes); n != 0 || err != io.ErrUnexpectedEOF {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, io.ErrUnexpectedEOF, n, err)
	}
}

func TestIndexSyncFailureIndexedPipe(t *testing.T) {
	testBytes := []byte("\nHello World\n\n")
	data := &mock.ReadWriteSeekable{}
	index := &mock.SyncReadWriteSeekable{ReadWriteSeekable: &mock.ReadWriteSeekable{}}
	p := bufpipe.NewLineIndexedPipe(data, index)
	index.SyncFunc = func() error {
		return io.ErrUnexpectedEOF
	}

	if n, err := p.Write(testBytes); n != 0 || err != io.ErrUnexpectedEOF {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, io.ErrUnexpectedEOF, n, err)
	}
}

func TestLateDataWriteFailIndexedPipe(t *testing.T) {
	testBytes := []byte("\nHello World\n\n")
	data := &mock.ReadWriteSeekable{}
	index := &mock.ReadWriteSeekable{}
	p := bufpipe.NewLineIndexedPipe(data, index)
	data.WriteFunc = unwriteableFunc

	if n, err := p.Write(testBytes); n != 0 || err != io.ErrUnexpectedEOF {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, io.ErrUnexpectedEOF, n, err)
	}
}

func TestIndexExtremeEdgeCaseFailureIndexedPipe(t *testing.T) {
	testBytes := []byte("\nHello World\n\n")
	data := &mock.ReadWriteSeekable{}
	index := &mock.ReadWriteSeekable{}
	p := bufpipe.NewLineIndexedPipe(data, index)

	data.SeekFunc = func(rws *mock.ReadWriteSeekable, offset int64, whence int) (int64, error) {
		fpcs := make([]uintptr, 1)
		runtime.Callers(3, fpcs)
		fun := runtime.FuncForPC(fpcs[0] - 1)
		if strings.HasSuffix(fun.Name(), "writeIndex") {
			return 0, errUnseekable
		}
		return 0, nil
	}

	if n, err := p.Write(testBytes); n != 0 || err != errUnseekable {
		t.Errorf("Expected [%v, %v] got [%v, %v]", 0, errUnseekable, n, err)
	}
}
