package core

import (
	"fmt"
	"runtime"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
)

// getProcessCPUUsage returns the CPU usage of the process
func getServiceCPUUsage() (float64, error) {
	return common.GetProcessObject().CPUPercent()
}

// getServiceGoroutines returns the number of goroutines in the service
func getServiceGoroutines() int {
	return runtime.NumGoroutine()
}

// calculateMemoryUsagePercentage calculates memory usage percentage
func calculateMemoryUsagePercentage(usedMemory, totalMemory string) (float64, error) {
	totalMemoryMB, err := common.ConvertToMB(totalMemory)
	if err != nil {
		return 0, err
	}
	usedMemoryMB, err := common.ConvertToMB(usedMemory)
	if err != nil {
		return 0, err
	}
	return (usedMemoryMB / totalMemoryMB) * 100, nil
}

// calculateServiceHealth calculates service health based on CPU, memory, and goroutines
func calculateServiceHealth(stats *models.ServiceStats) (float64, string, error) {
	cpuUsage, err := getServiceCPUUsage()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get service CPU usage: %w", err)
	}

	totalAvailableCores := stats.CPUStatistics.TotalCores
	cpuUsagePercentage := (cpuUsage / float64(totalAvailableCores)) * 100

	// Calculating memory usage percentage for the service
	memoryUsagePercentage, err := calculateMemoryUsagePercentage(
		stats.MemoryStatistics.MemoryUsedByService,
		stats.MemoryStatistics.TotalSystemMemory,
	)
	if err != nil {
		return 0, "", fmt.Errorf("failed to calculate memory usage percentage: %w", err)
	}

	// Calculating the health ratios for CPU, memory, and goroutines
	cpuUsageRatio := (cpuUsagePercentage / serviceHealthThresholds.MaxCPUUsage) * 100
	memoryUsageRatio := (memoryUsagePercentage / serviceHealthThresholds.MaxMemoryUsage) * 100
	goRoutinesRatio := (float64(getServiceGoroutines()) / float64(serviceHealthThresholds.MaxGoRoutines)) * 100
	finalScore := (cpuUsageRatio + memoryUsageRatio + goRoutinesRatio) / 3

	var message string
	if finalScore > 100 {
		finalScore = 100
		message = fmt.Sprintf(
			"Service usage exceeds allowed limits: CPU Usage %.2f%% / %.2f%%, Memory Usage %.2f%% / %.2f%%, Goroutines %.2f / %d",
			cpuUsageRatio, serviceHealthThresholds.MaxCPUUsage,
			memoryUsageRatio, serviceHealthThresholds.MaxMemoryUsage,
			goRoutinesRatio, serviceHealthThresholds.MaxGoRoutines,
		)
	} else {
		finalScore = 100 - finalScore
		message = fmt.Sprintf(
			"Service usage is within limits: CPU Usage %.2f%% / %.2f%%, Memory Usage %.2f%% / %.2f%%, Goroutines %.2f / %d",
			cpuUsageRatio, serviceHealthThresholds.MaxCPUUsage,
			memoryUsageRatio, serviceHealthThresholds.MaxMemoryUsage,
			goRoutinesRatio, serviceHealthThresholds.MaxGoRoutines,
		)
	}

	return finalScore, message, nil
}

// calculateSystemHealth calculates system health based on CPU and memory
func calculateSystemHealth(stats *models.ServiceStats) (float64, string, error) {

	// Calculating cpu & memory usage percentage for the system
	cpuUsagePercentage, err := GetCPUPrecent()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get CPU percent: %w", err)
	}
	memoryUsagePercentage, err := calculateMemoryUsagePercentage(
		stats.MemoryStatistics.MemoryUsedBySystem,
		stats.MemoryStatistics.TotalSystemMemory,
	)
	if err != nil {
		return 0, "", fmt.Errorf("failed to calculate memory usage percentage: %w", err)
	}

	cpuUsageRatio := (cpuUsagePercentage / serviceHealthThresholds.MaxCPUUsage) * 100
	memoryUsageRatio := (memoryUsagePercentage / serviceHealthThresholds.MaxMemoryUsage) * 100
	finalScore := (cpuUsageRatio + memoryUsageRatio) / 2
	var message string
	if finalScore > 100 {
		finalScore = 0
		message = fmt.Sprintf(
			"System usage exceeds allowed limits: CPU Usage %.2f%% / %.2f%%, Memory Usage %.2f%% / %.2f%%",
			cpuUsageRatio, serviceHealthThresholds.MaxCPUUsage,
			memoryUsageRatio, serviceHealthThresholds.MaxMemoryUsage,
		)
	} else {
		finalScore = 100 - finalScore
		message = fmt.Sprintf(
			"System usage is within limits: CPU Usage %.2f%% / %.2f%%, Memory Usage %.2f%% / %.2f%%",
			cpuUsageRatio, serviceHealthThresholds.MaxCPUUsage,
			memoryUsageRatio, serviceHealthThresholds.MaxMemoryUsage,
		)
	}

	return finalScore, message, nil
}

// CalculateHealthScore calculates the health score of both the system and service
func CalculateHealthScore(serviceStats *models.ServiceStats) (*models.SystemHealthInPercent, error) {
	// Calculating system health
	systemScore, systemMsg, err := calculateSystemHealth(serviceStats)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate system health: %w", err)
	}

	// CalcCalculating service health
	serviceScore, serviceMsg, err := calculateServiceHealth(serviceStats)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate service health: %w", err)
	}

	return &models.SystemHealthInPercent{
		SystemHealth: models.HealthFields{
			Percentage:    common.RoundFloat64(systemScore, 2),
			AllowedByUser: serviceHealthThresholds.MaxCPUUsage,
			Message:       systemMsg,
		},
		ServiceHealth: models.HealthFields{
			Percentage:    common.RoundFloat64(serviceScore, 2),
			AllowedByUser: serviceHealthThresholds.MaxCPUUsage,
			Message:       serviceMsg,
		},
	}, nil
}
