// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package bufpipe_test

import (
	"io/ioutil"
	"os"
	"syscall"
	"testing"

	"github.com/Ladbrokes/bufpipe"
)

func TestNewLineIndexedFilePipe(t *testing.T) {
	data, err := ioutil.TempFile("", "testdata")
	if err != nil {
		t.Error("Unable to create temporary data file", err)
	}
	index, err := ioutil.TempFile("", "testindex")
	if err != nil {
		t.Error("Unable to create temporary index file", err)
	}

	// Test open existing...
	p, err := bufpipe.NewLineIndexedFilePipe(data.Name(), index.Name(), 0666)
	if err != nil {
		t.Error("Unable to create IndexedFile object", err)
	}

	p.Close()

	os.Remove(data.Name())
	os.Remove(index.Name())

	// Test open new...
	p, err = bufpipe.NewLineIndexedFilePipe(data.Name(), index.Name(), 0666)
	if err != nil {
		t.Error("Unable to create IndexedFile object", err)
	}

	p.Close()

	os.Remove(data.Name())
	os.Remove(index.Name())

	// Test unopenable data
	os.Mkdir(data.Name(), 0666)
	p, err = bufpipe.NewLineIndexedFilePipe(data.Name(), index.Name(), 0666)
	if err != nil {
		if perr, ok := err.(*os.PathError); ok {
			if perr.Err != syscall.EISDIR {
				t.Errorf("Expected %v got %v", syscall.EISDIR, perr.Err)
			}
		} else {
			t.Errorf("Expected *os.PathError got %T", err)
		}
	}

	os.Remove(data.Name())
	os.Remove(index.Name())

	// Test unopenable index
	os.Mkdir(index.Name(), 0666)
	p, err = bufpipe.NewLineIndexedFilePipe(data.Name(), index.Name(), 0666)
	if err != nil {
		if perr, ok := err.(*os.PathError); ok {
			if perr.Err != syscall.EISDIR {
				t.Errorf("Expected %v got %v", syscall.EISDIR, perr.Err)
			}
		} else {
			t.Errorf("Expected *os.PathError got %T", err)
		}
	}

	os.Remove(data.Name())
	os.Remove(index.Name())

}
