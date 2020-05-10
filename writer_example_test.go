// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio_test

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/tunabay/go-speedio"
)

//
func ExampleWriter_devNull() {
	// /dev/null writer with rate limit of 256 bit/s
	w, err := speedio.NewWriter(ioutil.Discard, 256)
	if err != nil {
		panic(err)
	}

	// print bit rate every 0.5 seconds
	go func() {
		for range time.Tick(time.Second / 2) {
			fmt.Println("bitrate:", w.BitRate())
		}
	}()

	// start measurement
	w.Start()
	time.Sleep(time.Second) // measurement resolution

	// write 160 bytes data, it will take about 5 seconds
	buf := make([]byte, 160)
	n, err := w.Write(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println("written:", n)

	// close wrapper only (don't want to close /dev/null)
	if err := w.CloseSingle(); err != nil {
		panic(err)
	}

	// print total
	bc, et, br := w.Total()
	fmt.Println("total byte count:", bc)
	fmt.Println("elapsed time:", et)
	fmt.Println("total bit rate:", br)
}
