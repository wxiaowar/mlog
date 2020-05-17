package mlog

import (
	"log"
	"math"
	"sync/atomic"
	"unsafe"
)

type Alerter func(int)

type RingItem struct {
	Data []byte
	seq  uint64
}

type Ring struct {
	wIdx    uint64
	rIdx    uint64
	mask    uint64
	buffer  []unsafe.Pointer
	tryMax  int
	alerter Alerter
}

// NewRing creates a new ring buffer.
func NewRing(try, max int, alerter Alerter) *Ring {
	if alerter == nil {
		alerter = func(int) {}
	}

	// find 11111
	cnt := 0
	for max > 0 {
		cnt++
		max = max / 2
	}

	mask := uint64(1 << cnt - 1)
	d := &Ring{
		buffer:  make([]unsafe.Pointer, mask+1),
		mask:    mask,
		tryMax:  try,
		alerter: alerter,
	}

	//fmt.Printf("len = %b\n", len(d.buffer))
	d.wIdx = math.MaxUint64
	return d
}

// Add add the data in the next slot
func (r *Ring) Add(val *RingItem) {
	cnt := 0
	for {
		if cnt > r.tryMax {
			r.alerter(cnt) // alert
			break
		}

		wIdx := atomic.AddUint64(&r.wIdx, 1)
		idx := wIdx & r.mask
		ptr := atomic.LoadPointer(&(r.buffer[idx]))

		// when not nil, try next slot
		if ptr != nil && (*RingItem)(ptr) != nil {
			cnt++
			log.Printf("ring collision %d-%d: consider using a larger buf \n", cnt, idx)
			continue
		}

		val.seq = wIdx

		// multi add on one slot
		if !atomic.CompareAndSwapPointer(&r.buffer[idx], ptr, unsafe.Pointer(val)) {
			cnt++
			log.Printf("ring collision %d-%d: atomic\n", cnt, idx)
			continue
		}

		break
	}
}

// Next will attempt to read from the next slot of the ring buffer.
// If there is not data available, it will return (nil, false).
func (r *Ring) Next() (val *RingItem, ok bool) {
	idx := r.rIdx & r.mask
	val = (*RingItem)(atomic.SwapPointer(&r.buffer[idx], nil))
	if val == nil {
		return nil, false
	}

	r.rIdx++
	return val, true
}
