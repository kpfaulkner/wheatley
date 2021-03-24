package Metrics

import "time"

type AzureMetricName string

var ProcessorCpuPercentageMetric AzureMetricName = "performanceCounters/processorCpuPercentage"
var MemoryAvailableBytesMetric AzureMetricName = "performanceCounters/memoryAvailableBytes"

type CPUProcessorAveragePercentage struct {
	Value struct {
		Start        time.Time `json:"start"`
		End          time.Time `json:"end"`
		Interval     string    `json:"interval"`
		TimeSegments []struct {
			Start    time.Time `json:"start"`
			End      time.Time `json:"end"`
			Segments []struct {
				CPUPercentage struct {
					Avg float64 `json:"avg"`
				} `json:"performanceCounters/processorCpuPercentage"`
				CloudRoleName string `json:"cloud/roleName"`
			} `json:"segments"`
		} `json:"segments"`
	} `json:"value"`
}

type MemoryAvailableBytes struct {
	Value struct {
		Start        time.Time `json:"start"`
		End          time.Time `json:"end"`
		Interval     string    `json:"interval"`
		TimeSegments []struct {
			Start    time.Time `json:"start"`
			End      time.Time `json:"end"`
			Segments []struct {
				MemoryAvailable struct {
					Avg float64 `json:"avg"`
				} `json:"performanceCounters/memoryAvailableBytes"`
				CloudRoleName string `json:"cloud/roleName"`
			} `json:"segments"`
		} `json:"segments"`
	} `json:"value"`
}
