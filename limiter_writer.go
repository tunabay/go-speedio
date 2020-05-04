// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/tunabay/go-infounit"
)

// LimiterWriter implements bit rate limiting for an io.Writer object.
type LimiterWriter struct {
	wr         io.Writer // underlying writer provided by the client
	rate       infounit.BitRate
	resolution time.Duration
	maxWait    time.Duration
	lim        *limiter
	closed     bool
	closedChan chan struct{}
	mu         sync.RWMutex
}

// NewLimiterWriter creates a new LimiterWriter with default configuration. The
// default configuration has a resolution of 1s and a max-wait duration of 500ms,
// which means that average bit rate for the last 1s is limited, and a partial
// writing is performed at least every 500ms.
func NewLimiterWriter(wr io.Writer, rate infounit.BitRate) (*LimiterWriter, error) {
	return NewLimiterWriterWithConfig(wr, rate, nil)
}

// NewLimiterWriterWithConfig creates a new LimiterWriter with the specified
// configuration. If conf is nil, the default configuration will be used.
func NewLimiterWriterWithConfig(wr io.Writer, rate infounit.BitRate, conf *LimiterConfig) (*LimiterWriter, error) {
	if conf == nil {
		conf = DefaultLimiterConfig
	}
	w := &LimiterWriter{
		wr:         wr,
		rate:       rate,
		resolution: conf.Resolution,
		maxWait:    conf.MaxWait,
		closedChan: make(chan struct{}),
	}
	lim, err := newLimiter(w.rate, w.resolution, w.maxWait)
	if err != nil {
		return nil, err
	}
	w.lim = lim
	return w, nil
}

// LimitingBitRate returns the current limiting bit rate.
func (w *LimiterWriter) LimitingBitRate() infounit.BitRate {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.rate
}

// SetBitRate sets a new limiting bit rate.
func (w *LimiterWriter) SetBitRate(rate infounit.BitRate) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.lim.set(time.Now(), rate, w.resolution, w.maxWait); err != nil {
		return err
	}
	w.rate = rate
	return nil
}

// Close closes the writer.
// If the underlying writer implements io.WriteCloser, its Close method
// is also called.
func (w *LimiterWriter) Close() error {
	return w.close(true)
}

// CloseSingle is the same as Close except that it does not close the underlying writer.
func (w *LimiterWriter) CloseSingle() error {
	return w.close(false)
}

//
func (w *LimiterWriter) close(chain bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	close(w.closedChan)
	if !chain {
		return nil
	}
	if wrc, ok := w.wr.(io.Closer); ok {
		if err := wrc.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Write writes len(p) bytes from p to the underlying writer.
// It blocks until all the data in p is written, but it does not just wait
// until all the data is written. It repeatedly writes part of the divided p
// to the underlying writer.
func (w *LimiterWriter) Write(p []byte) (int, error) {
	written := 0
	for 0 < len(p) {
		tc := time.Now()
		wd, abc := w.lim.request(tc, len(p))
		// fmt.Printf("DEBUG: len=%d, wd=%s, allowed=%d, total=%d\n", len(p), wd, abc, written)
		if 0 < wd {
			timer := time.NewTimer(wd)
			select {
			case <-timer.C:
			case <-w.closedChan:
				if !timer.Stop() {
					<-timer.C
				}
				return written, fmt.Errorf("closed")
			}
		}
		n, err := w.wr.Write(p[:abc])
		if n < abc {
			w.lim.refund(abc - n)
		}
		if err != nil {
			return written, err
		}
		if n == 0 {
			return written, fmt.Errorf("zero bytes written")
		}
		written += n
		p = p[n:]
	}
	return written, nil
}
