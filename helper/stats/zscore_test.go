package stats

import (
	"context"
	"math"
	"testing"

	"pulse/internal/contract"
)

func TestZScore_InsufficientBaseline(t *testing.T) {
	calc := NewZScoreCalculator(5)
	tests := []struct {
		name     string
		baseline []float64
	}{
		{"nil", nil},
		{"empty", []float64{}},
		{"one", []float64{10}},
		{"four", []float64{10, 20, 30, 40}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := calc.ZScore(context.Background(), contract.ZScoreInput{
				Current:  100,
				Baseline: tt.baseline,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.ZScore != 0 {
				t.Errorf("expected ZScore=0 for insufficient data, got %f", r.ZScore)
			}
			if r.Sufficient {
				t.Error("expected Sufficient=false for insufficient data")
			}
		})
	}
}

func TestZScore_FlatBaseline(t *testing.T) {
	calc := NewZScoreCalculator(3)
	r, err := calc.ZScore(context.Background(), contract.ZScoreInput{
		Current:  50,
		Baseline: []float64{50, 50, 50, 50, 50},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ZScore != 0 {
		t.Errorf("expected ZScore=0 for flat baseline, got %f", r.ZScore)
	}
	if !r.Sufficient {
		t.Error("expected Sufficient=true for flat baseline")
	}
}

func TestZScore_KnownValue(t *testing.T) {
	// Baseline: [10, 20, 30, 40, 50]
	// Mean = 30, Stddev = sqrt(200) ≈ 14.1421
	// Current = 72.4264 → z = (72.4264 - 30) / 14.1421 ≈ 3.0
	calc := NewZScoreCalculator(5)
	baseline := []float64{10, 20, 30, 40, 50}

	// Compute expected z-score manually.
	m := 30.0
	sd := math.Sqrt(200.0)
	current := m + 3.0*sd // exactly 3 standard deviations above mean

	r, err := calc.ZScore(context.Background(), contract.ZScoreInput{
		Current:  current,
		Baseline: baseline,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Sufficient {
		t.Fatal("expected Sufficient=true")
	}
	if math.Abs(r.ZScore-3.0) > 1e-6 {
		t.Errorf("expected ZScore≈3.0, got %f", r.ZScore)
	}
}

func TestZScore_Negative(t *testing.T) {
	calc := NewZScoreCalculator(3)
	// Baseline: [100, 100, 100, 100, 100]
	// Mean = 100, stddev = 0 → but let's use varied baseline
	baseline := []float64{100, 110, 90, 105, 95}
	// Mean = 100, stddev = sqrt(50) ≈ 7.071
	// Current = 50 → z = (50 - 100) / 7.071 ≈ -7.07

	r, err := calc.ZScore(context.Background(), contract.ZScoreInput{
		Current:  50,
		Baseline: baseline,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ZScore >= 0 {
		t.Errorf("expected negative ZScore for below-mean current, got %f", r.ZScore)
	}
}

func TestZScore_DefaultMinBaseline(t *testing.T) {
	calc := NewZScoreCalculator(0) // should default to 5
	if calc.MinBaseline != 5 {
		t.Errorf("expected default MinBaseline=5, got %d", calc.MinBaseline)
	}

	calc2 := NewZScoreCalculator(-1) // negative should also default
	if calc2.MinBaseline != 5 {
		t.Errorf("expected default MinBaseline=5 for negative input, got %d", calc2.MinBaseline)
	}
}

func TestZScore_ExactMinBaseline(t *testing.T) {
	calc := NewZScoreCalculator(3)

	// Exactly 3 samples — should be sufficient.
	r, err := calc.ZScore(context.Background(), contract.ZScoreInput{
		Current:  100,
		Baseline: []float64{10, 20, 30},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Sufficient {
		t.Error("expected Sufficient=true with exactly MinBaseline samples")
	}
	if r.ZScore == 0 {
		t.Error("expected non-zero ZScore when current differs from mean")
	}
}
