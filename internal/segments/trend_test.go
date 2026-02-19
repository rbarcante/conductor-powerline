package segments

import "testing"

func TestTrendIncreasing(t *testing.T) {
	arrow := TrendArrow(50.0, 40.0)
	if arrow != "\u2191" { // ↑
		t.Errorf("expected up arrow, got %q", arrow)
	}
}

func TestTrendDecreasing(t *testing.T) {
	arrow := TrendArrow(30.0, 50.0)
	if arrow != "\u2193" { // ↓
		t.Errorf("expected down arrow, got %q", arrow)
	}
}

func TestTrendStableWithinThreshold(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		previous float64
	}{
		{"exactly same", 50.0, 50.0},
		{"within +2%", 51.5, 50.0},
		{"within -2%", 48.5, 50.0},
		{"at +2% boundary", 52.0, 50.0},
		{"at -2% boundary", 48.0, 50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrow := TrendArrow(tt.current, tt.previous)
			if arrow != "\u2192" { // →
				t.Errorf("expected stable arrow, got %q", arrow)
			}
		})
	}
}

func TestTrendNoPreviousData(t *testing.T) {
	arrow := TrendArrow(50.0, -1)
	if arrow != "" {
		t.Errorf("expected empty string for no previous data, got %q", arrow)
	}
}

func TestTrendZeroValues(t *testing.T) {
	arrow := TrendArrow(0.0, 0.0)
	if arrow != "\u2192" { // →
		t.Errorf("expected stable arrow for zero values, got %q", arrow)
	}
}
