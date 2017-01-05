// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package bufpipe

import (
	"io"
	"os"
)

// LineIndexedFilePipe LineIndexedPipe around Files
type LineIndexedFilePipe struct {
	*LineIndexedPipe
	data  string
	index string
}

// NewLineIndexedFilePipe will create and return a LineIndexedFilePipe based around the given
// data and index filenames.
// The files will be created if required with the given permissions, if the files already exist
// they will be opened for appending.
func NewLineIndexedFilePipe(data, index string, perm os.FileMode) (*LineIndexedFilePipe, error) {
	var err error
	l := &LineIndexedFilePipe{
		data:  data,
		index: index,
	}

	var dataFile, indexFile *os.File
	if dataFile, err = os.OpenFile(data, os.O_APPEND|os.O_CREATE|os.O_RDWR, perm); err != nil {
		return nil, err
	}

	if indexFile, err = os.OpenFile(index, os.O_APPEND|os.O_CREATE|os.O_RDWR, perm); err != nil {
		dataFile.Close()
		return nil, err
	}

	l.LineIndexedPipe = NewLineIndexedPipe(dataFile, indexFile)

	return l, nil
}

// Close closes the Pipe, rendering then unusable for I/O. It returns an error, if any.
func (l *LineIndexedFilePipe) Close() error {
	l.LineIndexedPipe.data.(io.Closer).Close()
	l.LineIndexedPipe.index.(io.Closer).Close()

	l.LineIndexedPipe.close()

	return nil
}
