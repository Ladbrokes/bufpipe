// Copyright 2017 Shannon Wynter, Ladbrokes Digital Australia Pty Ltd. All rights reserved.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

package bufpipe

import (
	"bytes"
	"testing"
)

func TestScanLines(t *testing.T) {
	// End of file
	if adv, token, err := scanLines([]byte{}, true); adv != 0 || token != nil || err != nil {
		t.Errorf("Expected [%v, %v, %v] got [%v, %v, %v]", 0, nil, nil, adv, token, err)
	}

	// End of file but still have data to process
	if adv, token, err := scanLines([]byte("H\n"), true); adv != 2 || !bytes.Equal([]byte{'H', '\n'}, token) || err != nil {
		t.Errorf("Expected [%v, %v, %v] got [%v, %v, %v]", 2, []byte{'H', '\n'}, nil, adv, token, err)
	}

	// End of file but still have data to process, no trailing newline
	if adv, token, err := scanLines([]byte("H"), true); adv != 1 || !bytes.Equal([]byte{'H'}, token) || err != nil {
		t.Errorf("Expected [%v, %v, %v] got [%v, %v, %v]", 1, []byte{'H'}, nil, adv, token, err)
	}

	// Not end of file but no data given
	if adv, token, err := scanLines([]byte{}, false); adv != 0 || token != nil || err != nil {
		t.Errorf("Expected [%v, %v, %v] got [%v, %v, %v]", 0, nil, nil, adv, token, err)
	}

}
