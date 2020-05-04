// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"io"
	"time"

	"github.com/tunabay/go-infounit"
)

// Writer implements both bit rate limiting and bit rate measurement for an io.Writer object.
// If only one of limiting or measurement is needed, it is recommended to use
// LimiterWriter or MeterWriter instead.
// In fact, Writer is just a concatenation of LimiterWriter and MeterWriter.
type Writer struct {
	mw *MeterWriter
	lw *LimiterWriter
}

// NewWriter creates a new Writer with default configurations.
func NewWriter(wr io.Writer, rate infounit.BitRate) (*Writer, error) {
	return NewWriterWithConfig(wr, rate, nil, nil)
}

// NewWriterWithConfig creates a new Writer with the specified
// configurations. If conf is nil, the default configuration will be used.
func NewWriterWithConfig(wr io.Writer, rate infounit.BitRate, lconf *LimiterConfig, mconf *MeterConfig) (*Writer, error) {
	mw, err := NewMeterWriterWithConfig(wr, mconf)
	if err != nil {
		return nil, err
	}
	lw, err := NewLimiterWriterWithConfig(mw, rate, lconf)
	if err != nil {
		return nil, err
	}
	return &Writer{mw: mw, lw: lw}, nil
}

// LimitingBitRate returns the current effective limiting bit rate.
func (w *Writer) LimitingBitRate() infounit.BitRate {
	return w.lw.LimitingBitRate()
}

// SetBitRate sets a new bit rate limiting.
func (w *Writer) SetBitRate(rate infounit.BitRate) error {
	return w.lw.SetBitRate(rate)
}

// Close closes the writer.
// If the underlying writer implements io.WriteCloser, its Close method
// is also called.
func (w *Writer) Close() error {
	return w.lw.Close()
}

// CloseAt is the same as Close, except that it uses time specified as the end
// time.
func (w *Writer) CloseAt(tc time.Time) error {
	if err := w.mw.CloseAt(tc); err != nil {
		_ = w.lw.CloseSingle()
		return err
	}
	return w.lw.CloseSingle()
}

// CloseSingle is the same as Close except that it does not close the underlying writer.
func (w *Writer) CloseSingle() error {
	if err := w.mw.CloseSingle(); err != nil {
		_ = w.lw.CloseSingle()
		return err
	}
	return w.lw.CloseSingle()
}

// CloseSingleAt is the same as CloseAt except that it does not close the underlying writer.
func (w *Writer) CloseSingleAt(tc time.Time) error {
	if err := w.mw.CloseSingleAt(tc); err != nil {
		_ = w.lw.CloseSingle()
		return err
	}
	return w.lw.CloseSingle()
}

// Write writes len(p) bytes from p to the underlying writer.
// It blocks until all the data in p is written, but it does not just wait
// until all the data is written. It repeatedly writes part of the divided p
// to the underlying writer.
func (w *Writer) Write(p []byte) (int, error) {
	return w.lw.Write(p)
}

// Start starts the measurement. Calling this Start is optional, and normally it
// is started automatically at the first write. This is used to adjust the
// transfer start time for bit rate calculation.
func (w *Writer) Start() {
	w.mw.Start()
}

// StartAt starts the measurement at specified time. This is used to adjust the
// transfer start time for bit rate calculation.
func (w *Writer) StartAt(tc time.Time) {
	w.mw.StartAt(tc)
}

// BitRate calculates and returns the bit rate in the most recent sampling
// period.
func (w *Writer) BitRate() infounit.BitRate {
	return w.mw.BitRate()
}

// Total returns the data transfer amount, elapsed time, and bit rate in the
// entire period from start. When it is called after being closed, it always
// returns the same statistics from start to close.
func (w *Writer) Total() (infounit.ByteCount, time.Duration, infounit.BitRate) {
	return w.mw.Total()
}
