package service

import (
	"math"
	"sort"
)

// calculateMean calculates the arithmetic mean of a slice of values
func (s *AnalysisService) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// calculateMedian calculates the median of a slice of values
func (s *AnalysisService) calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// calculateStdDev calculates the standard deviation of a slice of values
func (s *AnalysisService) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	sumSquares := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(values)-1)
	return math.Sqrt(variance)
}

// calculatePercentile calculates the nth percentile of a sorted slice using linear interpolation
func (s *AnalysisService) calculatePercentile(sortedValues []float64, percentile int) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	if percentile <= 0 {
		return sortedValues[0]
	}
	if percentile >= 100 {
		return sortedValues[len(sortedValues)-1]
	}

	// Calculate the fractional index within the sorted slice.
	index := float64(percentile) / 100.0 * float64(len(sortedValues)-1)
	lowerIndex := int(math.Floor(index))
	upperIndex := int(math.Ceil(index))

	if lowerIndex == upperIndex {
		return sortedValues[lowerIndex]
	}

	// Linear interpolation between the two values
	weight := index - float64(lowerIndex)
	return sortedValues[lowerIndex]*(1-weight) + sortedValues[upperIndex]*weight
}

// calculatePercentileFromValues calculates the percentile ranking of a value within a dataset
func (s *AnalysisService) calculatePercentileFromValues(value float64, values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Count how many values are less than or equal to the target value
	count := 0
	for _, v := range values {
		if v <= value {
			count++
		}
	}

	// Calculate percentile (percentage of values that are <= target)
	return float64(count) / float64(len(values)) * 100
}

// normalCDF approximates the cumulative distribution function of the standard normal distribution
func (s *AnalysisService) normalCDF(z float64) float64 {
	// Abramowitz and Stegun approximation
	if z < -8 {
		return 0
	}
	if z > 8 {
		return 1
	}

	sum := 0.0
	term := z
	i := 3.0
	for math.Abs(term) > 1e-10 && i < 100 {
		sum += term
		term *= z * z / i
		i += 2
	}

	return 0.5 + sum*0.3989422804014327 // 1/sqrt(2*pi)
}
