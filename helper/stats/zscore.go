package stats

import (
	"context"
	"math"

	"pulse/internal/contract"
)

const defaultMinBaseline = 5

var _ contract.StatsProvider = (*ZScoreCalculator)(nil)

// ZScoreCalculator computes rolling z-scores from pre-aggregated window data.
// When the baseline has fewer than MinBaseline samples, it returns a zero score
// with Sufficient=false so callers can skip anomaly detection on cold starts.
type ZScoreCalculator struct {
	// MinBaseline is the minimum number of baseline buckets required for a
	// statistically reliable z-score. Baselines shorter than this produce
	// ZScoreResult{ZScore: 0, Sufficient: false}.
	MinBaseline int
}

// NewZScoreCalculator creates a calculator with the given minimum baseline size.
// If minBaseline <= 0, the default of 5 is used.
func NewZScoreCalculator(minBaseline int) *ZScoreCalculator {
	if minBaseline <= 0 {
		minBaseline = defaultMinBaseline
	}
	return &ZScoreCalculator{MinBaseline: minBaseline}
}

// ZScore computes the z-score for the current observation against the baseline.
//
// Formula: z = (current − mean) / stddev
//
// Special cases:
//   - Baseline length < MinBaseline → {ZScore: 0, Sufficient: false}
//   - Standard deviation ≈ 0 (flat baseline) → {ZScore: 0, Sufficient: true}
func (c *ZScoreCalculator) ZScore(_ context.Context, input contract.ZScoreInput) (contract.ZScoreResult, error) {
	if len(input.Baseline) < c.MinBaseline {
		return contract.ZScoreResult{ZScore: 0, Sufficient: false}, nil
	}

	mean := mean(input.Baseline)
	sd := stddev(input.Baseline, mean)

	if sd < 1e-9 {
		return contract.ZScoreResult{ZScore: 0, Sufficient: true}, nil
	}

	z := (input.Current - mean) / sd
	return contract.ZScoreResult{ZScore: z, Sufficient: true}, nil
}

// mean returns the arithmetic mean of vals.
func mean(vals []float64) float64 {
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// stddev returns the population standard deviation given a precomputed mean.
func stddev(vals []float64, m float64) float64 {
	var sumSq float64
	for _, v := range vals {
		d := v - m
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(vals)))
}
