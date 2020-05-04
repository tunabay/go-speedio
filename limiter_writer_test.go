// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio_test

import (
	"io/ioutil"
	"testing"

	"github.com/tunabay/go-speedio"
)

//
func TestLimiterWriter_test1(t *testing.T) {
	t.Parallel()

	w, err := speedio.NewLimiterWriter(ioutil.Discard, 2048)
	if err != nil {
		t.Error(err)
		return
	}
	buf := make([]byte, 1024)
	n, err := w.Write(buf)
	if n != 1024 {
		t.Errorf("unexpected len: want: 1024, got: %d", n)
		return
	}
	if err != nil {
		t.Error(err)
	}
	if err := w.Close(); err != nil {
		t.Error(err)
	}
}

//
func TestLimiterWriter_test2(t *testing.T) {
	t.Parallel()

	if _, err := speedio.NewLimiterWriter(ioutil.Discard, 1); err == nil {
		t.Errorf("error expected")
	}
}
