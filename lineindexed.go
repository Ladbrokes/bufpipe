// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package bufpipe

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// LineIndexedPipe provides a ReadWriter interface to store line deliminited data in an indexed fasion in a pair of ReadWriteSeekers
type LineIndexedPipe struct {
	*Pipe
	index io.ReadWriteSeeker

	lastIndex int64
}

const int64Size = 8

// NewLineIndexedPipe returns a new line indexed pipe structure
func NewLineIndexedPipe(data, index io.ReadWriteSeeker) *LineIndexedPipe {
	l := &LineIndexedPipe{
		Pipe:  NewPipe(data),
		index: index,
	}
	return l
}

// Write implements the standard Write interface: it writes data to the pipe, blocking
// until readers have consumed all the data or the read end is closed. If the read end
// is closed with an error, that err is returned as err; otherwise err is ErrClosedPipe.
func (l *LineIndexedPipe) Write(p []byte) (n int, err error) {
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

		// Go back to last known good size
		if _, err = l.data.Seek(l.size, os.SEEK_SET); err != nil {
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(p))
		scanner.Split(scanLines)
		for scanner.Scan() {
			var wn int

			output := scanner.Bytes()
			wn, err = l.data.Write(output)
			n += wn

			if err != nil {
				return
			}

			if output[len(output)-1] == '\n' {
				if err = l.writeIndex(); err != nil {
					n = n - wn
					return
				}
			}

			// Succeeded in writing the index so this is the new file length
			l.size += int64(wn)
		}

		err = scanner.Err()

		l.rwait.Signal()

		return
	}()
	if err != nil {
		l.werr = err
	}
	return
}

// SeekLine sets the reader position to the beginning of the given line
func (l *LineIndexedPipe) SeekLine(line int64) (err error) {
	err = func() (err error) {

		l.l.Lock()
		defer l.l.Unlock()

		if _, err = l.index.Seek(line*int64Size, os.SEEK_SET); err != nil {
			return
		}

		err = binary.Read(l.index, binary.LittleEndian, &l.readIndex)

		return
	}()
	if err != nil {
		l.rerr = err
	}
	return
}

// CountLines returns the number of lines stored
func (l *LineIndexedPipe) CountLines() (int64, error) {
	size, err := l.IndexSize()
	return size / int64Size, err
}

// IndexSize returns the size of the index on disk/in memory
func (l *LineIndexedPipe) IndexSize() (int64, error) {
	l.l.Lock()
	defer l.l.Unlock()

	return l.index.Seek(0, os.SEEK_END)
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0 : i+1], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (l *LineIndexedPipe) writeIndex() (err error) {
	if _, err = l.index.Seek(0, os.SEEK_END); err != nil {
		return
	}

	if err = binary.Write(l.index, binary.LittleEndian, l.lastIndex); err != nil {
		return err
	}

	if err = l.syncWriter(l.index); err != nil {
		return err
	}

	if l.lastIndex, err = l.data.Seek(0, os.SEEK_END); err != nil {
		return err
	}

	return
}
