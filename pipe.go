// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package bufpipe

import (
	"errors"
	"io"
	"os"
	"sync"
)

// Pipe provides a ReadWriter interface to store data in a ReadWriteSeeker
type Pipe struct {
	data io.ReadWriteSeeker

	readIndex int64
	size      int64

	l sync.Mutex // protects remaining fields

	rwait sync.Cond // waiting reader

	rerr error // if reader closed, error to give writes
	werr error // if writer closed, error to give reads

}

// ErrBadSeekOffset is returned when seek doesn't end up at the expected location
var ErrBadSeekOffset = errors.New("seek offset failure")

// Syncer interface lists pipe optionally sync changes
type Syncer interface {
	Sync() error
}

// NewPipe returns a new line indexed pipe structure
func NewPipe(data io.ReadWriteSeeker) *Pipe {
	size, err := data.Seek(0, os.SEEK_END)
	if err != nil {
		panic(err)
	}

	l := &Pipe{
		data: data,
		size: size,
	}

	l.rwait.L = &l.l
	return l
}

// Read implements the standard Read interface: it reads data from the pipe, blocking
// until a writer arrives or the write end is closed. If the write end is closed with
// an error, that error is returned as err; otherwise err is EOF.
func (l *Pipe) Read(d []byte) (n int, err error) {
	n, err = func() (n int, err error) {
		l.l.Lock()
		defer l.l.Unlock()

		for {
			if l.rerr != nil {
				return 0, io.ErrClosedPipe
			}
			if l.werr != nil {
				return 0, l.werr
			}
			if l.readIndex < l.size {
				break
			}
			l.rwait.Wait()
		}

		if _, err = l.data.Seek(l.readIndex, os.SEEK_SET); err != nil {
			return
		}

		n, err = l.data.Read(d)
		l.readIndex += int64(n)
		return
	}()
	if err != nil {
		l.rerr = err
	}
	return
}

// Write implements the standard Write interface: it writes data to the pipe, blocking
// until readers have consumed all the data or the read end is closed. If the read end
// is closed with an error, that err is returned as err; otherwise err is ErrClosedPipe.
func (l *Pipe) Write(d []byte) (n int, err error) {
	n, err = func() (n int, err error) {
		l.l.Lock()
		defer l.l.Unlock()

		if l.rerr != nil {
			err = l.rerr
			return
		}
		if l.werr != nil {
			err = io.ErrClosedPipe
			return
		}

		if _, err = l.data.Seek(0, os.SEEK_END); err != nil {
			return
		}

		n, err = l.data.Write(d)

		if err == nil {
			err = l.syncWriter(l.data)
		}

		l.size += int64(n)
		l.rwait.Signal()

		return
	}()
	if err != nil {
		l.werr = err
	}
	return
}

// Seek sets the offset for the next Read to offset, the Write offset is always the end of the data.
func (l *Pipe) Seek(offset int64, whence int) (int64, error) {
	l.l.Lock()
	defer l.l.Unlock()

	switch whence {
	case os.SEEK_END:
		l.readIndex = l.size - offset
	case os.SEEK_SET:
		l.readIndex = offset
	case os.SEEK_CUR:
		l.readIndex += offset
	}

	if l.readIndex > l.size {
		l.readIndex = l.size
	}
	if l.readIndex < 0 {
		l.readIndex = 0
	}

	return l.readIndex, nil
}

// DataSize returns the size of the data stored on disk/in memory
func (l *Pipe) DataSize() (int64, error) {
	l.l.Lock()
	defer l.l.Unlock()

	return l.size, nil
}

func (l *Pipe) close() {
	l.l.Lock()
	defer l.l.Unlock()
	l.rerr = io.ErrClosedPipe
	l.werr = io.ErrClosedPipe
	l.rwait.Signal()
}

func (l *Pipe) syncWriter(w io.Writer) error {
	if t, ok := w.(Syncer); ok {
		return t.Sync()
	}
	return nil
}
