// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/tunabay/go-speedio"
)

//
func TestMeterWriter_test1(t *testing.T) {
	t.Parallel()

	w := speedio.NewMeterWriter(ioutil.Discard)

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.Log("bitrate:", w.BitRate())
			case <-done:
				return
			}
		}
	}()

	buf := make([]byte, 123000/8/5)
	sum := 0

	w.Start()

	ticker := time.NewTicker(time.Second / 5)
	for {
		<-ticker.C
		n, err := w.Write(buf)
		if 0 < n {
			sum += n
			if 123000/8*5 <= sum {
				break
			}
		}
		if err != nil {
			t.Error(err)
			break
		}
		if n != len(buf) {
			t.Errorf("unexpected len: want: %d, got: %d", len(buf), n)
		}
	}
	ticker.Stop()
	if err := w.Close(); err != nil {
		t.Error(err)
	}
	close(done)
	bc, et, br := w.Total()
	t.Logf("total(n=%d): %v, %v, %v", sum, bc, et, br)
}
