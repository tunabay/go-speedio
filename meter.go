// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/tunabay/go-infounit"
)

// meter measures the latest data transfer amount.
type meter struct {
	resolution          time.Duration
	sample              time.Duration
	cur, last, first    *meterItem
	started, closed     bool
	startedAt, closedAt time.Time
	totalBytes          infounit.ByteCount
	mu                  sync.RWMutex
}

// meterItem is an element of linked list for the meter.
type meterItem struct {
	vol        float64
	start      time.Duration // must be a multiple of resolution
	end        time.Time
	next, prev *meterItem
}

// newMeter creates a meter with specified resolution and sample duration.
func newMeter(resolution, sample time.Duration) (*meter, error) {
	switch {
	case resolution <= 0:
		return nil, fmt.Errorf("%w: resolution %d <= 0", ErrInvalidParameter, resolution)
	case resolution < MinMeterResolution:
		return nil, fmt.Errorf("%w: resolution %d < minMeterResolution", ErrInvalidParameter, resolution)
	case sample <= 0:
		return nil, fmt.Errorf("%w: sample %d <= 0", ErrInvalidParameter, sample)
	}
	n := int(sample / resolution)
	if n < 2 {
		return nil, fmt.Errorf("%w: too small sample duration %s (at least %s)", ErrInvalidParameter, sample, resolution*2)
	}
	m := &meter{
		resolution: resolution,
		sample:     sample,
		cur:        &meterItem{},
	}

	tail := m.cur
	for i := 0; i < n; i++ {
		tail.next = &meterItem{prev: tail}
		tail = tail.next
	}
	m.cur.prev = tail
	tail.next = m.cur

	return m, nil
}

// start starts measuring the data transfer.
func (m *meter) start(tc time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return
	}
	m.started, m.startedAt = true, tc
	m.cur.start, m.cur.end = 0, m.startedAt.Add(m.resolution)
}

// close stops measuring the data transfer.
func (m *meter) close(tc time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	m.closed, m.closedAt = true, tc
}

// record records the data transfer into the meter.
func (m *meter) record(tc time.Time, b infounit.ByteCount) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tc.Before(m.cur.end) {
		m.cur.vol += float64(b)
		m.totalBytes += b
		return
	}
	m.last, m.cur = m.cur, m.cur.next
	switch m.first {
	case nil:
		m.first = m.last
	case m.cur:
		m.first = m.first.next
	}
	m.cur.start = tc.Sub(m.startedAt) / m.resolution * m.resolution
	m.cur.end = m.startedAt.Add(m.cur.start + m.resolution)
	m.cur.vol = float64(b)
	m.totalBytes += b
}

// bpscoef is a coefficient used for bit rate calculation.
const bpscoef = 8 * float64(time.Second)

// bitRate returns the bit rate in the last sample period.
func (m *meter) bitRate(tc time.Time) infounit.BitRate {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.last == nil {
		return infounit.BitRate(0)
	}

	newest := m.last
	sampleEnd := m.cur.start
	if !tc.Before(m.cur.end) { // tc is after the cur
		newest = m.cur
		sampleEnd = tc.Sub(m.startedAt) / m.resolution * m.resolution
	}
	sampleStart := sampleEnd - m.sample
	sampleWidth := m.sample
	if sampleStart < 0 {
		sampleStart = 0
		sampleWidth = sampleEnd
	}

	var sum float64
	for i := newest; sampleStart <= i.start; i = i.prev {
		sum += i.vol
		if i == m.first {
			break
		}
	}
	// fmt.Printf("DEBUG: %f * 8 / %s\n", sum, sampleWidth)
	return infounit.BitRate(sum * bpscoef / float64(sampleWidth))
}

// total returns the data transfer amount, elapsed time, and bit rate
// in the entire period from start to close.
func (m *meter) total(tc time.Time) (infounit.ByteCount, time.Duration, infounit.BitRate) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.last == nil {
		return 0, 0, 0
	}
	if m.closed {
		tc = m.closedAt
	}
	b, d := m.totalBytes, tc.Sub(m.startedAt)
	switch {
	case b == 0:
		return 0, d, 0
	case d == 0:
		return b, d, infounit.BitRate(math.Inf(+1))
	}
	return b, d, infounit.BitRate(float64(b) * bpscoef / float64(d))
}
