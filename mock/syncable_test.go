// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package mock_test

import (
	"errors"
	"testing"

	"github.com/Ladbrokes/bufpipe/mock"
)

var errMockMe = errors.New("Mock me")

func TestSyncReadWriteSeekable(t *testing.T) {
	works := false
	srws := mock.SyncReadWriteSeekable{
		SyncFunc: func() error {
			works = true
			return nil
		},
	}

	if err := srws.Sync(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if works != true {
		t.Error("Expected callback to be called, was not called")
	}
}

func TestSyncableReadWriterSeekableError(t *testing.T) {
	works := false
	srws := mock.SyncReadWriteSeekable{
		SyncFunc: func() error {
			works = true
			return errMockMe
		},
	}

	if err := srws.Sync(); err != errMockMe {
		t.Errorf("Expected %v got %v", errMockMe, err)
	}

	if works != true {
		t.Error("Expected callback to be called, was not called")

	}
}

func TestSyncReadWriteSeekableNoFunc(t *testing.T) {
	works := false
	srws := mock.SyncReadWriteSeekable{}

	if err := srws.Sync(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if works == true {
		t.Error("Expected callback not to be called, was called")
	}
}
