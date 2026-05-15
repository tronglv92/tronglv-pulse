package stats

import (
	"math"
	"testing"
)

func TestEWMA_FirstValueSetsAverage(t *testing.T) {
	e := NewEWMA(0.5)
	if e.Ready() {
		t.Error("expected Ready()=false before any Add")
	}

	e.Add(100)
	if !e.Ready() {
		t.Error("expected Ready()=true after Add")
	}
	if e.Value() != 100 {
		t.Errorf("expected first value=100, got %f", e.Value())
	}
}

func TestEWMA_Smoothing(t *testing.T) {
	e := NewEWMA(0.5) // α = 0.5
	e.Add(100)        // avg = 100
	e.Add(200)        // avg = 0.5*200 + 0.5*100 = 150
	if e.Value() != 150 {
		t.Errorf("expected 150, got %f", e.Value())
	}

	e.Add(200) // avg = 0.5*200 + 0.5*150 = 175
	if e.Value() != 175 {
		t.Errorf("expected 175, got %f", e.Value())
	}
}

func TestEWMA_HighAlpha_TracksRecent(t *testing.T) {
	e := NewEWMA(0.9)
	e.Add(0)
	e.Add(100)
	// avg = 0.9*100 + 0.1*0 = 90
	if math.Abs(e.Value()-90) > 1e-9 {
		t.Errorf("expected ~90, got %f", e.Value())
	}
}

func TestEWMA_LowAlpha_SmoothsHeavily(t *testing.T) {
	e := NewEWMA(0.1)
	e.Add(0)
	e.Add(100)
	// avg = 0.1*100 + 0.9*0 = 10
	if math.Abs(e.Value()-10) > 1e-9 {
		t.Errorf("expected ~10, got %f", e.Value())
	}
}

func TestEWMA_Reset(t *testing.T) {
	e := NewEWMA(0.5)
	e.Add(100)
	e.Add(200)
	e.Reset()

	if e.Ready() {
		t.Error("expected Ready()=false after Reset")
	}
	if e.Value() != 0 {
		t.Errorf("expected Value()=0 after Reset, got %f", e.Value())
	}

	// First add after reset should set directly.
	e.Add(50)
	if e.Value() != 50 {
		t.Errorf("expected 50 after Reset+Add, got %f", e.Value())
	}
}

func TestEWMA_AlphaClamping(t *testing.T) {
	// Alpha <= 0 should be clamped to smallest positive float.
	e := NewEWMA(0)
	e.Add(100)
	e.Add(200)
	// With near-zero alpha, average barely moves from 100.
	if math.Abs(e.Value()-100) > 1 {
		t.Errorf("expected ~100 with near-zero alpha, got %f", e.Value())
	}

	// Alpha > 1 should be clamped to 1.
	e2 := NewEWMA(5.0)
	e2.Add(100)
	e2.Add(200)
	// With alpha=1, average always equals the latest value.
	if e2.Value() != 200 {
		t.Errorf("expected 200 with alpha=1, got %f", e2.Value())
	}
}

func TestEWMA_Convergence(t *testing.T) {
	e := NewEWMA(0.3)
	e.Add(0)
	// Feed constant 100 many times — should converge to 100.
	for i := 0; i < 100; i++ {
		e.Add(100)
	}
	if math.Abs(e.Value()-100) > 0.01 {
		t.Errorf("expected convergence to ~100, got %f", e.Value())
	}
}

func TestEWMA_ZeroValue(t *testing.T) {
	e := NewEWMA(0.5)
	if e.Value() != 0 {
		t.Errorf("expected 0 before any Add, got %f", e.Value())
	}
}
