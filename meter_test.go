// Copyright (c) 2020 Hirotsuna Mizuno. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package speedio_test

import (
	"testing"
	"time"

	"github.com/tunabay/go-infounit"
	"github.com/tunabay/go-speedio"
)

//
func TestMeter_test1(t *testing.T) {
	t.Parallel()

	rec, err := speedio.ExportNewMeter(time.Second, time.Second*3)
	if err != nil {
		t.Errorf("newMeter: %s", err)
		return
	}

	tm := time.Now()

	speedio.ExportMeterStart(rec, tm)
	for i := 0; i < 10; i++ {
		t.Logf("==== %dsec ====", i)
		speedio.ExportMeterRecord(rec, tm.Add(time.Millisecond*time.Duration(1000*i+100)), 1000+infounit.ByteCount(250))
		speedio.ExportMeterRecord(rec, tm.Add(time.Millisecond*time.Duration(1000*i+200)), 1000+infounit.ByteCount(250))
		speedio.ExportMeterRecord(rec, tm.Add(time.Millisecond*time.Duration(1000*i+300)), 1000+infounit.ByteCount(250))
		speedio.ExportMeterRecord(rec, tm.Add(time.Millisecond*time.Duration(1000*i+990)), 1000+infounit.ByteCount(250))
		t.Logf("%v", speedio.ExportMeterBitRate(rec, tm.Add(time.Millisecond*time.Duration(1000*i+990))))
		rec.DebugDump()
	}
	speedio.ExportMeterRecord(rec, tm.Add(time.Millisecond*time.Duration(1000*10+100)), 1000+infounit.ByteCount(4000))
	speedio.ExportMeterClose(rec, tm.Add(time.Second*11))
	bc, et, br := speedio.ExportMeterTotal(rec, tm.Add(time.Second*100))
	t.Logf("TOTAL: %v, %v, %v", bc, et, br)
	rec.DebugDump()
}
