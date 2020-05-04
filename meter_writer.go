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

// MeterWriter implements bit rate measurement for an io.Writer object.
type MeterWriter struct {
	wr              io.Writer // underlying writer provided by the client
	resolution      time.Duration
	sample          time.Duration
	met             *meter
	started, closed bool
	mu              sync.Mutex
}

// NewMeterWriter creates a new MeterWriter with default configuration. The
// default configuration has a resolution of 500ms and a sample duration of 3s,
// which means that average bit rate for the last 3s is updated every 500ms.
func NewMeterWriter(wr io.Writer) *MeterWriter {
	w, err := NewMeterWriterWithConfig(wr, nil)
	if err != nil {
		panic(err)
	}
	return w
}

// NewMeterWriterWithConfig creates a new MeterWriter with the specified
// configuration. If conf is nil, the default configuration will be used.
func NewMeterWriterWithConfig(wr io.Writer, conf *MeterConfig) (*MeterWriter, error) {
	if conf == nil {
		conf = DefaultMeterConfig
	}
	w := &MeterWriter{
		wr:         wr,
		resolution: conf.Resolution,
		sample:     conf.Sample,
	}
	met, err := newMeter(w.resolution, w.sample)
	if err != nil {
		return nil, err
	}
	w.met = met
	return w, nil
}

// Start starts the measurement. Calling this Start is optional, and normally it
// is started automatically at the first write. This is used to adjust the
// transfer start time for bit rate calculation.
func (w *MeterWriter) Start() {
	w.StartAt(time.Now())
}

// StartAt starts the measurement at specified time. This is used to adjust the
// transfer start time for bit rate calculation.
func (w *MeterWriter) StartAt(tc time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.started {
		return
	}
	w.started = true
	w.met.start(tc)
}

// Close closes the writer, which means that it ends the bit rate calculation
// period. If the underlying writer implements io.WriteCloser, its Close method
// is also called.
func (w *MeterWriter) Close() error {
	return w.close(time.Now(), true)
}

// CloseAt is the same as Close, except that it uses time specified as the end
// time.
func (w *MeterWriter) CloseAt(tc time.Time) error {
	return w.close(tc, true)
}

// CloseSingle is the same as Close except that it does not close the underlying writer.
func (w *MeterWriter) CloseSingle() error {
	return w.close(time.Now(), false)
}

// CloseSingleAt is the same as CloseAt except that it does not close the underlying writer.
func (w *MeterWriter) CloseSingleAt(tc time.Time) error {
	return w.close(tc, false)
}

//
func (w *MeterWriter) close(tc time.Time, chain bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	if !w.started {
		w.started = true
		w.met.start(tc)
	}
	w.met.close(tc)
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

// Write writes len(p) bytes from p to the underlying writer. It is caller's
// responsibility to call Close after writing all data to record the measurement
// end time.
func (w *MeterWriter) Write(p []byte) (int, error) {
	w.Start()
	n, err := w.wr.Write(p)
	if 0 < n {
		w.met.record(time.Now(), infounit.ByteCount(n))
	}
	return n, err
}

// BitRate calculates and returns the bit rate in the most recent sampling
// period.
func (w *MeterWriter) BitRate() infounit.BitRate {
	return w.met.bitRate(time.Now())
}

// Total returns the data transfer amount, elapsed time, and bit rate in the
// entire period from start. When it is called after being closed, it always
// returns the same statistics from start to close.
func (w *MeterWriter) Total() (infounit.ByteCount, time.Duration, infounit.BitRate) {
	return w.met.total(time.Now())
}
