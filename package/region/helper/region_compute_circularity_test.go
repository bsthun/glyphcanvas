package regionHelper

import (
	"math"
	"testing"
)

func TestRegionComputeCircularity(t *testing.T) {
	tests := []struct {
		name     string
		hu       []float64
		expected float64
		epsilon  float64
	}{
		{
			name:     "High circularity Hu moments",
			hu:       []float64{0.16, 0.0001},
			expected: 0.9412,
			epsilon:  0.01,
		},
		{
			name:     "Low circularity Hu moments",
			hu:       []float64{0.20, 0.005},
			expected: 0.7388,
			epsilon:  0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RegionComputeCircularity(tt.hu)
			if math.Abs(result-tt.expected) > tt.epsilon {
				t.Errorf("RegionComputeCircularity() = %v, want %v (±%v)", result, tt.expected, tt.epsilon)
			}
		})
	}
}

func BenchmarkRegionComputeCircularity(b *testing.B) {
	hu := []float64{0.16, 0.0001}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RegionComputeCircularity(hu)
	}
}
