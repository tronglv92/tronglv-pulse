package contract

import "context"

// ZScoreInput holds the pre-aggregated window data needed to compute a z-score.
type ZScoreInput struct {
	// Current is the observed count in the latest time bucket.
	Current float64
	// Baseline is the historical bucket counts used to compute mean and stddev.
	Baseline []float64
}

// ZScoreResult carries the computed z-score and a data-sufficiency flag.
type ZScoreResult struct {
	ZScore float64
	// Sufficient is false when Baseline has too few buckets for a reliable score.
	// Callers must not trigger anomaly events when Sufficient is false.
	Sufficient bool
}

// StatsProvider computes statistical signals over the log stream.
type StatsProvider interface {
	// ZScore computes a rolling z-score from pre-aggregated window data.
	ZScore(ctx context.Context, input ZScoreInput) (ZScoreResult, error)
}
