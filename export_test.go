// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

// Bridge package to expose speedio internals to tests in the speedio_test
// package.

package speedio

import (
	"fmt"
)

//
type ExportMeter = meter

//
var (
	ExportNewMeter     = newMeter
	ExportMeterStart   = (*meter).start
	ExportMeterClose   = (*meter).close
	ExportMeterRecord  = (*meter).record
	ExportMeterBitRate = (*meter).bitRate
	ExportMeterTotal   = (*meter).total
)

//
func (r *MeterReader) DebugDump() {
	r.met.DebugDump()
}

//
func (m *meter) DebugDump() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fmt.Printf("METER:\n")
	for i := m.cur; ; i = i.prev {
		tag := ""
		if i == m.cur {
			tag += " [cur]"
		}
		if i == m.last {
			tag += " [last]"
		}
		if i == m.first {
			tag += " [first]"
		}
		fmt.Printf("%p: s=%s, e=%s, vol=%v%s\n", i, i.start, i.end, i.vol, tag)
		if i == m.first || m.last == nil {
			break
		}
	}
}
