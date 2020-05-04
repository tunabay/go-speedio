// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio_test

import (
	"io"
	"testing"

	"github.com/tunabay/go-randdata"
	"github.com/tunabay/go-speedio"
)

//
func TestLimiterReader_test1(t *testing.T) {
	t.Parallel()

	rd := randdata.New(randdata.Binary, 0, 5)
	r, err := speedio.NewLimiterReader(rd, 16)
	if err != nil {
		t.Error(err)
		return
	}
	buf := make([]byte, 1024)
	sum := 0
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
	}
	r.Close()
	if sum != 5 {
		t.Errorf("unexpected data len: want=5, got=%d", sum)
	}
}

//
func TestLimiterReader_test2(t *testing.T) {
	t.Parallel()

	rd := randdata.New(randdata.Binary, 0, 5)
	_, err := speedio.NewLimiterReader(rd, 1)
	if err == nil {
		t.Errorf("error expected")
		return
	}
}
