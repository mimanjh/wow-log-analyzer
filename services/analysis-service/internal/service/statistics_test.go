package service

import (
	"testing"
)

func TestAnalysisService_calculateMean(t *testing.T) {
	service := NewAnalysisService()

	tests := []struct {
		input    []float64
		expected float64
	}{
		{[]float64{1, 2, 3, 4, 5}, 3.0},
		{[]float64{10}, 10.0},
		{[]float64{}, 0.0},
		{[]float64{1.5, 2.5, 3.5}, 2.5},
	}

	for _, test := range tests {
		result := service.calculateMean(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected %f, got %f", test.input, test.expected, result)
		}
	}
}

func TestAnalysisService_calculateMedian(t *testing.T) {
	service := NewAnalysisService()

	tests := []struct {
		input    []float64
		expected float64
	}{
		{[]float64{1, 2, 3, 4, 5}, 3.0},
		{[]float64{1, 2, 3, 4}, 2.5},
		{[]float64{10}, 10.0},
		{[]float64{}, 0.0},
		{[]float64{1, 3, 5, 7, 9}, 5.0},
	}

	for _, test := range tests {
		result := service.calculateMedian(test.input)
		if result != test.expected {
			t.Errorf("For input %v, expected %f, got %f", test.input, test.expected, result)
		}
	}
}

func TestAnalysisService_calculatePercentile(t *testing.T) {
	service := NewAnalysisService()

	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		percentile int
		expected   float64
	}{
		{0, 1.0},
		{25, 3.25},
		{50, 5.5},
		{75, 7.75},
		{100, 10.0},
	}

	for _, test := range tests {
		result := service.calculatePercentile(values, test.percentile)
		if result != test.expected {
			t.Errorf("For percentile %d, expected %f, got %f", test.percentile, test.expected, result)
		}
	}
}

func TestAnalysisService_calculatePercentileFromValues(t *testing.T) {
	service := NewAnalysisService()

	values := []float64{1, 2, 3, 4, 5}

	tests := []struct {
		value    float64
		expected float64
	}{
		{1, 20.0},  // 1 value <= 1 out of 5 = 20%
		{2, 40.0},  // 2 values <= 2 out of 5 = 40%
		{3, 60.0},  // 3 values <= 3 out of 5 = 60%
		{5, 100.0}, // 5 values <= 5 out of 5 = 100%
		{6, 100.0}, // 5 values <= 6 out of 5 = 100%
		{0, 0.0},   // 0 values <= 0 out of 5 = 0%
	}

	for _, test := range tests {
		result := service.calculatePercentileFromValues(test.value, values)
		if result != test.expected {
			t.Errorf("For value %f, expected %f, got %f", test.value, test.expected, result)
		}
	}
}
