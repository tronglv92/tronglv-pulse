package stats

import "math"

// EWMA computes an Exponential Weighted Moving Average.
//
// Each call to Add updates the average:
//
//	ewma = α × value + (1 − α) × ewma
//
// A higher α (closer to 1) makes the average respond faster to recent values.
// A lower α (closer to 0) smooths out noise and weights history more heavily.
type EWMA struct {
	alpha float64
	value float64
	init  bool
}

// NewEWMA creates an EWMA with the given smoothing factor.
// Alpha is clamped to the (0, 1] range.
func NewEWMA(alpha float64) *EWMA {
	alpha = math.Max(math.SmallestNonzeroFloat64, math.Min(alpha, 1.0))
	return &EWMA{alpha: alpha}
}

// Add feeds a new observation into the moving average.
// The first observation sets the average directly (no smoothing).
func (e *EWMA) Add(value float64) {
	if !e.init {
		e.value = value
		e.init = true
		return
	}
	e.value = e.alpha*value + (1-e.alpha)*e.value
}

// Value returns the current moving average.
// Returns 0 if no observations have been added.
func (e *EWMA) Value() float64 {
	return e.value
}

// Ready reports whether at least one observation has been added.
func (e *EWMA) Ready() bool {
	return e.init
}

// Reset clears the state so the next Add sets the average directly again.
func (e *EWMA) Reset() {
	e.value = 0
	e.init = false
}
