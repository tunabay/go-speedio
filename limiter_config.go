// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"time"
)

// LimiterConfig indicates the configuration parameter of bit rate limiting.
//
// Resolution is the period for totaling the transfer amount to determine whether the bit rate is exceeded or not.
// For example, if the bit rate is 1 kbit/s and Resolution is 3s,
// if there is no transfer in the previous 2 seconds, transfer of 3 kbit is allowed
// in the next 1 second. However, when Resolution is 1s,
// the transfer allowed per second is always 1 kbit.
//
// MaxWait is the maximum waiting time when the transfer exceeds the bit rate.
// After this MaxWait time elapses, only the portion that is allowed at that time is transferred.
type LimiterConfig struct {
	Resolution time.Duration
	MaxWait    time.Duration
}

// DefaultLimiterConfig is the default configuration for bit rate limiting
// with 1s resolution and 500ms max-wait time.
// It means that the amount of data transferred in every 1s is limited,
// and the allowed amount of partial data is transferred every 500ms.
var DefaultLimiterConfig = &LimiterConfig{
	Resolution: time.Second,
	MaxWait:    time.Millisecond * 500,
}
