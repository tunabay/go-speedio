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

// MeterReader implements bit rate measurement for an io.Reader object.
type MeterReader struct {
	rd              io.Reader // underlying reader provided by the client
	resolution      time.Duration
	sample          time.Duration
	met             *meter
	started, closed bool
	mu              sync.Mutex
}

// NewMeterReader creates a new MeterReader with default configuration. The
// default configuration has a resolution of 500ms and a sample duration of 3s,
// which means that average bit rate for the last 3s is updated every 500ms.
func NewMeterReader(rd io.Reader) *MeterReader {
	r, err := NewMeterReaderWithConfig(rd, nil)
	if err != nil {
		panic(err)
	}
	return r
}

// NewMeterReaderWithConfig creates a new MeterReader with the specified
// configuration. If conf is nil, the default configuration will be used.
func NewMeterReaderWithConfig(rd io.Reader, conf *MeterConfig) (*MeterReader, error) {
	if conf == nil {
		conf = DefaultMeterConfig
	}
	r := &MeterReader{
		rd:         rd,
		resolution: conf.Resolution,
		sample:     conf.Sample,
	}
	met, err := newMeter(r.resolution, r.sample)
	if err != nil {
		return nil, err
	}
	r.met = met
	return r, nil
}

// Start starts the measurement. Calling this Start is optional, and normally it
// is started automatically at the first read. This is used to adjust the
// transfer start time for bit rate calculation.
func (r *MeterReader) Start() {
	r.StartAt(time.Now())
}

// StartAt starts the measurement at specified time. This is used to adjust the
// transfer start time for bit rate calculation.
func (r *MeterReader) StartAt(tc time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return
	}
	r.started = true
	r.met.start(tc)
}

// Close closes the reader, which means that it ends the bit rate calculation
// period. If the underlying reader implements io.ReadCloser, its Close method
// is also called.
func (r *MeterReader) Close() error {
	return r.close(time.Now(), true)
}

// CloseAt is the same as Close, except that it uses time specified as the end
// time.
func (r *MeterReader) CloseAt(tc time.Time) error {
	return r.close(tc, true)
}

// CloseSingle is the same as Close except that it does not close the underlying reader.
func (r *MeterReader) CloseSingle() error {
	return r.close(time.Now(), false)
}

// CloseSingleAt is the same as CloseAt except that it does not close the underlying reader.
func (r *MeterReader) CloseSingleAt(tc time.Time) error {
	return r.close(tc, false)
}

//
func (r *MeterReader) close(tc time.Time, chain bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	if !r.started {
		r.started = true
		r.met.start(tc)
	}
	r.met.close(tc)
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

// Read reads data from the underlying reader into p. io.EOF does not
// automatically stop the measurement. It is caller's responsibility to call
// Close after receiving io.EOF to record the measurement end time.
func (r *MeterReader) Read(p []byte) (int, error) {
	r.Start()
	n, err := r.rd.Read(p)
	if 0 < n {
		r.met.record(time.Now(), infounit.ByteCount(n))
	}
	return n, err
}

// BitRate calculates and returns the bit rate in the most recent sampling
// period.
func (r *MeterReader) BitRate() infounit.BitRate {
	return r.met.bitRate(time.Now())
}

// Total returns the data transfer amount, elapsed time, and bit rate in the
// entire period from start. When it is called after being closed, it always
// returns the same statistics from start to close.
func (r *MeterReader) Total() (infounit.ByteCount, time.Duration, infounit.BitRate) {
	return r.met.total(time.Now())
}
