// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/tunabay/go-randdata"
	"github.com/tunabay/go-speedio"
)

//
func TestWriter_test1(t *testing.T) {
	// t.Parallel()

	src := randdata.New(randdata.Binary, 0, 50000)
	w, err := speedio.NewWriter(ioutil.Discard, 80000)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("set rate:", w.LimitingBitRate())

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

	w.Start()
	time.Sleep(time.Second * 1) // limiter's resolution

	n, err := src.WriteTo(w)
	if err != nil {
		t.Error(err)
		return
	}
	if err := w.CloseSingle(); err != nil {
		t.Error(err)
	}
	close(done)

	bc, et, br := w.Total()
	t.Logf("total(n=%d): %v, %v, %v", n, bc, et, br)
}
