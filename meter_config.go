// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio

import (
	"time"
)

// MeterConfig indicates the configuration parameter of bit rate measurement.
//
// Resolution is how often the bitrate is updated.
// Shorter resolutions increase measurement overhead and memory usage.
// Longer resolutions increase measurement delay.
// For example, with a 10s resolution, the bit rate is 0 for the first 10 seconds.
//
// Sample is the length of the most recent period for which the simple moving average bit rate is calculated.
// It must be an integral multiple of Resolution for accurate measurements.
// Also, it must be at least twice the Resolution.
// Longer sample periods increase memory usage for measurements.
type MeterConfig struct {
	Resolution time.Duration
	Sample     time.Duration // must be an integral multiple of Resolution
}

// MinResolution is the minimum time resolution to measure bit rate.
const MinMeterResolution time.Duration = time.Millisecond * 100

// DefaultMeterConfig is the default configuration for measurement
// with 500ms resolution and 3s sample duration.
// It means that the average bitrate for the last 3s is updated every 500ms.
var DefaultMeterConfig = &MeterConfig{
	Resolution: time.Millisecond * 500,
	Sample:     time.Second * 3,
}
