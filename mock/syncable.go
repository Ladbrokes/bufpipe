// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package mock

// SyncReadWriteSeekable provides a wrapper for the basic ReadWriteSeekable mock object to
// implement the syncer interface
type SyncReadWriteSeekable struct {
	*ReadWriteSeekable

	SyncFunc func() error
}

// Sync implements the syncer interface, returns an error if your callback returns an error.
func (srws *SyncReadWriteSeekable) Sync() error {
	if srws.SyncFunc != nil {
		return srws.SyncFunc()
	}
	return nil
}
