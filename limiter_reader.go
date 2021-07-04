// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"io"
	"sync"
	"time"

	"github.com/tunabay/go-infounit"
)

// LimiterReader implements bit rate limiting for an io.Reader object.
type LimiterReader struct {
	rd         io.Reader // underlying reader provided by the client
	rate       infounit.BitRate
	resolution time.Duration
	maxWait    time.Duration
	lim        *limiter
	closed     bool
	closedChan chan struct{}
	mu         sync.RWMutex
}

// NewLimiterReader creates a new LimiterReader with default configuration. The
// default configuration has a resolution of 1s and a max-wait duration of
// 500ms, which means that average bit rate for the last 1s is limited, and a
// partial reading is performed at least every 500ms.
func NewLimiterReader(rd io.Reader, rate infounit.BitRate) (*LimiterReader, error) {
	return NewLimiterReaderWithConfig(rd, rate, nil)
}

// NewLimiterReaderWithConfig creates a new LimiterReader with the specified
// configuration. If conf is nil, the default configuration will be used.
func NewLimiterReaderWithConfig(rd io.Reader, rate infounit.BitRate, conf *LimiterConfig) (*LimiterReader, error) {
	if conf == nil {
		conf = DefaultLimiterConfig
	}
	r := &LimiterReader{
		rd:         rd,
		rate:       rate,
		resolution: conf.Resolution,
		maxWait:    conf.MaxWait,
		closedChan: make(chan struct{}),
	}
	lim, err := newLimiter(r.rate, r.resolution, r.maxWait)
	if err != nil {
		return nil, err
	}
	r.lim = lim
	return r, nil
}

// LimitingBitRate returns the current limiting bit rate.
func (r *LimiterReader) LimitingBitRate() infounit.BitRate {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.rate
}

// SetBitRate sets a new limiting bit rate.
func (r *LimiterReader) SetBitRate(rate infounit.BitRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.lim.set(time.Now(), rate, r.resolution, r.maxWait); err != nil {
		return err
	}
	r.rate = rate
	return nil
}

// Close closes the reader. If the underlying reader implements io.ReadCloser,
// its Close method is also called.
func (r *LimiterReader) Close() error {
	return r.close(true)
}

// CloseSingle is the same as Close except that it does not close the underlying reader.
func (r *LimiterReader) CloseSingle() error {
	return r.close(false)
}

//
func (r *LimiterReader) close(chain bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	close(r.closedChan)
	if !chain {
		return nil
	}
	if rdc, ok := r.rd.(io.Closer); ok {
		if err := rdc.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Read reads data from the underlying reader into p. This may return shorter
// length than len(p). It may block for up to maxWait time.
func (r *LimiterReader) Read(p []byte) (int, error) {
	tc := time.Now()
	wd, abc := r.lim.request(tc, len(p))
	if 0 < wd {
		timer := time.NewTimer(wd)
		select {
		case <-timer.C:
		case <-r.closedChan:
			if !timer.Stop() {
				<-timer.C
			}
			return 0, ErrClosed
		}
	}
	n, err := r.rd.Read(p[:abc])
	if n < abc {
		r.lim.refund(abc - n)
	}
	return n, err
}
