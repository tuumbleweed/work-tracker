package worktracker

import (
	"math"
	"sync/atomic"
	"time"
)

// this is what we save to the JSONL file
type Chunk struct {
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at"`
	ActiveTime time.Duration `json:"active_time"`
}

// ActivityDurationTotals holds aggregated time values in milliseconds.
type ActivityDurationTotals struct {
	TotalDurationMs  int64   `json:"total_duration_ms"`  // total duration for today
	ActiveMsWeighted float64 `json:"active_ms_weighted"` // total amount of active ms
}

type AtomicFloat64 struct{ bits uint64 }

func (f *AtomicFloat64) Load() float64   { return math.Float64frombits(atomic.LoadUint64(&f.bits)) }
func (f *AtomicFloat64) Store(v float64) { atomic.StoreUint64(&f.bits, math.Float64bits(v)) }
func (f *AtomicFloat64) Add(d float64) float64 {
	for {
		oldBits := atomic.LoadUint64(&f.bits)
		old := math.Float64frombits(oldBits)
		nv := old + d
		if atomic.CompareAndSwapUint64(&f.bits, oldBits, math.Float64bits(nv)) {
			return nv
		}
	}
}
