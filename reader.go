// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"io"
	"time"

	"github.com/tunabay/go-infounit"
)

// Reader implements both bit rate limiting and bit rate measurement for an io.Reader object.
// If only one of limiting or measurement is needed, it is recommended to use
// LimiterReader or MeterReader instead.
// In fact, Reader is just a concatenation of LimiterReader and MeterReader.
type Reader struct {
	mr *MeterReader
	lr *LimiterReader
}

// NewReader creates a new Reader with default configurations.
func NewReader(rd io.Reader, rate infounit.BitRate) (*Reader, error) {
	return NewReaderWithConfig(rd, rate, nil, nil)
}

// NewReaderWithConfig creates a new Reader with the specified
// configurations. If conf is nil, the default configuration will be used.
func NewReaderWithConfig(rd io.Reader, rate infounit.BitRate, lconf *LimiterConfig, mconf *MeterConfig) (*Reader, error) {
	mr, err := NewMeterReaderWithConfig(rd, mconf)
	if err != nil {
		return nil, err
	}
	lr, err := NewLimiterReaderWithConfig(mr, rate, lconf)
	if err != nil {
		return nil, err
	}
	return &Reader{mr: mr, lr: lr}, nil
}

// LimitingBitRate returns the current effective limiting bit rate.
func (w *Reader) LimitingBitRate() infounit.BitRate {
	return w.lr.LimitingBitRate()
}

// SetBitRate sets a new bit rate limiting.
func (w *Reader) SetBitRate(rate infounit.BitRate) error {
	return w.lr.SetBitRate(rate)
}

// Close closes the reader.
// If the underlying reader implements io.ReadCloser, its Close method
// is also called.
func (w *Reader) Close() error {
	return w.lr.Close()
}

// CloseAt is the same as Close, except that it uses time specified as the end
// time.
func (w *Reader) CloseAt(tc time.Time) error {
	if err := w.mr.CloseAt(tc); err != nil {
		_ = w.lr.CloseSingle()
		return err
	}
	return w.lr.CloseSingle()
}

// CloseSingle is the same as Close except that it does not close the underlying reader.
func (w *Reader) CloseSingle() error {
	if err := w.mr.CloseSingle(); err != nil {
		_ = w.lr.CloseSingle()
		return err
	}
	return w.lr.CloseSingle()
}

// CloseSingleAt is the same as CloseAt except that it does not close the underlying reader.
func (w *Reader) CloseSingleAt(tc time.Time) error {
	if err := w.mr.CloseSingleAt(tc); err != nil {
		_ = w.lr.CloseSingle()
		return err
	}
	return w.lr.CloseSingle()
}

// Read reads data from the underlying reader into p. This may return shorter length
// than len(p). It may block for up to maxWait time.
// io.EOF does not automatically stop the measurement. It is caller's responsibility
// to call Close after receiving io.EOF to record the measurement end time.
func (w *Reader) Read(p []byte) (int, error) {
	return w.lr.Read(p)
}

// Start starts the measurement. Calling this Start is optional, and normally it
// is started automatically at the first read. This is used to adjust the
// transfer start time for bit rate calculation.
func (w *Reader) Start() {
	w.mr.Start()
}

// StartAt starts the measurement at specified time. This is used to adjust the
// transfer start time for bit rate calculation.
func (w *Reader) StartAt(tc time.Time) {
	w.mr.StartAt(tc)
}

// BitRate calculates and returns the bit rate in the most recent sampling
// period.
func (w *Reader) BitRate() infounit.BitRate {
	return w.mr.BitRate()
}

// Total returns the data transfer amount, elapsed time, and bit rate in the
// entire period from start. When it is called after being closed, it always
// returns the same statistics from start to close.
func (w *Reader) Total() (infounit.ByteCount, time.Duration, infounit.BitRate) {
	return w.mr.Total()
}
