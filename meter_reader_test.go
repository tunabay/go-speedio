// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio_test

import (
	"io"
	"testing"
	"time"

	"github.com/tunabay/go-randdata"
	"github.com/tunabay/go-speedio"
)

//
func TestMeterReader_test1(t *testing.T) {
	t.Parallel()

	src := randdata.New(randdata.Binary, 0, 500000)
	r := speedio.NewMeterReader(src)

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.Log("bitrate:", r.BitRate())
			case <-done:
				return
			}
		}
	}()

	buf := make([]byte, 20000)
	sum := 0

	r.Start()

	ticker := time.NewTicker(time.Second / 5)
	for {
		<-ticker.C
		n, err := io.ReadFull(r, buf)
		if 0 < n {
			sum += n
		}
		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				t.Error(err)
			}
			break
		}
		if n < 1 {
			t.Errorf("unexpected n: %d", n)
			break
		}
	}
	ticker.Stop()
	if err := r.Close(); err != nil {
		t.Error(err)
	}
	close(done)
	bc, et, br := r.Total()
	t.Logf("total(n=%d): %v, %v, %v", sum, bc, et, br)
}
