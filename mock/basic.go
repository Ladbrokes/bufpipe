// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package mock

import (
	"bytes"
	"io"
	"os"
)

// ReadWriteSeekable is a mock object loosely based on bytes.Buffer but with seeking behaviour comparable to os.File
type ReadWriteSeekable struct {
	buf []byte
	off int64

	// Override default functions
	ReadFunc  func(rws *ReadWriteSeekable, b []byte) (n int, err error)
	WriteFunc func(rws *ReadWriteSeekable, b []byte) (n int, err error)
	SeekFunc  func(rws *ReadWriteSeekable, offset int64, whence int) (int64, error)
	LenFunc   func(rws *ReadWriteSeekable) int
}

// NewReadWriteSeekable returns a ReadWriteSeekable with a pre-populated buffer
func NewReadWriteSeekable(buf []byte) *ReadWriteSeekable {
	return &ReadWriteSeekable{
		buf: buf,
	}
}

// Read reads up to len(b) bytes from the buffer.
// It returns the number of bytes read and an error, if any.
// EOF is signaled by a zero count with err set to io.EOF.
func (rws *ReadWriteSeekable) Read(b []byte) (n int, err error) {
	if rws.ReadFunc != nil {
		return rws.ReadFunc(rws, b)
	}

	if rws.buf == nil || rws.off >= int64(len(rws.buf)) {
		return 0, io.EOF
	}
	n = copy(b, rws.buf[rws.off:])
	rws.off += int64(n)
	return
}

func (rws *ReadWriteSeekable) grow(n int) {
	bl := len(rws.buf)
	if bl+n > cap(rws.buf) {
		defer func() {
			if recover() != nil {
				panic(bytes.ErrTooLarge)
			}
		}()

		if rws.buf == nil {
			rws.buf = make([]byte, 64)
		} else {
			rws.buf = append(rws.buf, make([]byte, n*2)...)
		}
	}
	rws.buf = rws.buf[0 : bl+n]
}

// Write writes len(b) bytes to the buffer.
// It returns the number of bytes written and an error, if any.
// Write returns a non-nil error when n != len(b).
func (rws *ReadWriteSeekable) Write(b []byte) (n int, err error) {
	if rws.WriteFunc != nil {
		return rws.WriteFunc(rws, b)
	}

	r := int64(len(rws.buf)) - rws.off
	if l := int64(len(b)); r < l {
		d := l - r
		rws.grow(int(d))
	}

	return copy(rws.buf[rws.off:], b), nil
}

// Seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means
// relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
func (rws *ReadWriteSeekable) Seek(offset int64, whence int) (int64, error) {
	if rws.SeekFunc != nil {
		return rws.SeekFunc(rws, offset, whence)
	}

	l := int64(len(rws.buf))
	switch whence {
	case os.SEEK_END:
		rws.off = l - offset
	case os.SEEK_SET:
		rws.off = offset
	case os.SEEK_CUR:
		rws.off += offset
	}

	if rws.off > l {
		rws.off = l
	}
	if rws.off < 0 {
		rws.off = 0
	}

	return int64(rws.off), nil
}

// Len returns the length of the buffer
func (rws *ReadWriteSeekable) Len() int {
	if rws.LenFunc != nil {
		return rws.LenFunc(rws)
	}

	return len(rws.buf)
}

// Reset resets the buffer and the offset to 0
func (rws *ReadWriteSeekable) Reset() {
	rws.off = 0
	rws.buf = nil

	rws.ReadFunc = nil
	rws.WriteFunc = nil
	rws.SeekFunc = nil
	rws.LenFunc = nil
}

// Bytes returns the contents of the buffer
func (rws *ReadWriteSeekable) Bytes() []byte {
	return rws.buf
}
