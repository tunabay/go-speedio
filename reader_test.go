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
func TestReader_test1(t *testing.T) {
	// t.Parallel()

	src := randdata.New(randdata.Binary, 0, 500000)
	r, err := speedio.NewReader(src, 800000)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("set rate:", r.LimitingBitRate())

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

	buf := make([]byte, 8192)
	sum := 0

	r.Start()
	time.Sleep(time.Second) // limiter's resolution

	for {
		n, err := r.Read(buf)
		if 0 < n {
			sum += n
		}
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		}
		if n < 1 {
			t.Errorf("unexpected n=%d", n)
			break
		}
	}
	if err := r.Close(); err != nil {
		t.Error(err)
	}
	close(done)
	bc, et, br := r.Total()
	t.Logf("total(n=%d): %v, %v, %v", sum, bc, et, br)
}
