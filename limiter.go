// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"fmt"
	"sync"
	"time"

	"github.com/tunabay/go-infounit"
)

// limiter limits the transfer.
type limiter struct {
	rate       float64 // bytes per sec ( = bps / 8 )
	burst      float64 // bytes
	minPartial int
	rateCoef   float64 // time.Second / rate
	lastTime   time.Time
	lastToken  float64 // bytes
	mu         sync.RWMutex
}

// resolution is the period for totaling the transfer amount to determine whether the bit rate is exceeded or not.
// For example, if the bit rate is 1 kbit/s and the resolution is 3s,
// if there is no transfer in the previous 2 seconds, transfer of 3 kbit is allowed
// in the next 1 second. However, when the resolution is 1s,
// the transfer allowed per second is always 1 kbit.
//
// maxWait is the maximum waiting time when the transfer exceeds the bit rate.
// After this maxWait time elapses, only the portion that can be transferred at that time is transferred.
func newLimiter(rate infounit.BitRate, resolution, maxWait time.Duration) (*limiter, error) {
	l := &limiter{}
	if err := l.set(time.Time{}, rate, resolution, maxWait); err != nil {
		return nil, err
	}
	return l, nil
}

//
func (l *limiter) set(tc time.Time, rate infounit.BitRate, resolution, maxWait time.Duration) error {
	switch {
	case rate < 0:
		return fmt.Errorf("negative bit rate %v", rate)
	case rate == 0:
		return fmt.Errorf("zero bit rate")
	case resolution < 0:
		return fmt.Errorf("negative resolution %s", resolution)
	case resolution == 0:
		return fmt.Errorf("zero resolution")
	case maxWait < 0:
		return fmt.Errorf("negative max-wait %s", maxWait)
	case maxWait == 0:
		return fmt.Errorf("zero max-wait")
	}

	newRate := float64(rate) / 8
	newBurst := newRate * resolution.Seconds()
	newMinPartial := int(newRate * maxWait.Seconds())

	switch {
	case infounit.ByteCount(newBurst) < 1:
		return fmt.Errorf("rate and/or resolution is too small: rate=%v, reso=%s", rate, resolution)
	case newMinPartial < 1:
		return fmt.Errorf("rate and/or max-wait is too small: rate=%v, wait=%s", rate, maxWait)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.lastTime.IsZero() && !tc.IsZero() {
		l.lastTime = tc
		l.lastToken += tc.Sub(l.lastTime).Seconds() * l.rate
		if l.burst < l.lastToken {
			l.lastToken = l.burst
		}
	}
	l.rate = newRate
	l.burst = newBurst
	l.minPartial = newMinPartial
	l.rateCoef = float64(time.Second) / l.rate

	return nil
}

// refund returns not used token.
func (l *limiter) refund(bc int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lastToken += float64(bc)
}

// request requests a transfer of the specified number of bytes.
// It returns the duration to wait and the number of bytes allowed.
func (l *limiter) request(tc time.Time, bc int) (time.Duration, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	allowed := l.lastToken + l.rate*tc.Sub(l.lastTime).Seconds()
	if l.burst < allowed {
		allowed = l.burst
	}
	allowedBytes := int(allowed)

	switch {
	case bc <= allowedBytes:
		l.lastTime = tc
		l.lastToken = allowed - float64(bc)
		return 0, bc
	case l.minPartial <= allowedBytes:
		l.lastTime = tc
		l.lastToken = allowed - float64(allowedBytes)
		return 0, allowedBytes
	}

	wsz := l.minPartial
	if bc < wsz {
		wsz = bc
	}
	d := time.Duration(l.rateCoef * (float64(wsz) - allowed))
	l.lastTime = tc.Add(d)
	l.lastToken = 0
	return d, wsz
}
