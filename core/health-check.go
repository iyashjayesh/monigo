package core

import (
	"fmt"
	"runtime"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/internal/alerting"
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
	if totalMemoryMB == 0 {
		return 0, fmt.Errorf("total memory is zero, cannot calculate usage percentage")
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
	t := GetThresholds()
	cpuUsageRatio := (cpuUsagePercentage / t.MaxCPUUsage) * 100
	memoryUsageRatio := (memoryUsagePercentage / t.MaxMemoryUsage) * 100
	goRoutinesRatio := (float64(getServiceGoroutines()) / float64(t.MaxGoRoutines)) * 100
	finalScore := (cpuUsageRatio + memoryUsageRatio + goRoutinesRatio) / 3

	var healthScore float64
	var message string
	if finalScore > 100 {
		healthScore = 0
		message = fmt.Sprintf(
			"Service usage exceeds allowed limits: CPU Usage %.2f%% / %.2f%%, Memory Usage %.2f%% / %.2f%%, Goroutines %.2f / %d",
			cpuUsageRatio, t.MaxCPUUsage,
			memoryUsageRatio, t.MaxMemoryUsage,
			goRoutinesRatio, t.MaxGoRoutines,
		)
	} else {
		healthScore = 100 - finalScore
		message = fmt.Sprintf(
			"Service usage is within limits: CPU Usage %.2f%% / %.2f%%, Memory Usage %.2f%% / %.2f%%, Goroutines %.2f / %d",
			cpuUsageRatio, t.MaxCPUUsage,
			memoryUsageRatio, t.MaxMemoryUsage,
			goRoutinesRatio, t.MaxGoRoutines,
		)
	}

	return healthScore, message, nil
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

	t := GetThresholds()
	cpuUsageRatio := (cpuUsagePercentage / t.MaxCPUUsage) * 100
	memoryUsageRatio := (memoryUsagePercentage / t.MaxMemoryUsage) * 100
	finalScore := (cpuUsageRatio + memoryUsageRatio) / 2

	var healthScore float64
	var message string
	if finalScore > 100 {
		healthScore = 0
		message = fmt.Sprintf(
			"System usage exceeds allowed limits: CPU Usage %.2f%% / %.2f%%, Memory Usage %.2f%% / %.2f%%",
			cpuUsageRatio, t.MaxCPUUsage,
			memoryUsageRatio, t.MaxMemoryUsage,
		)
	} else {
		healthScore = 100 - finalScore
		message = fmt.Sprintf(
			"System usage is within limits: CPU Usage %.2f%% / %.2f%%, Memory Usage %.2f%% / %.2f%%",
			cpuUsageRatio, t.MaxCPUUsage,
			memoryUsageRatio, t.MaxMemoryUsage,
		)
	}

	return healthScore, message, nil
}

// CalculateHealthScore calculates the health score of both the system and service
func CalculateHealthScore(serviceStats *models.ServiceStats) (*models.SystemHealthInPercent, error) {
	// Calculating system health
	systemScore, systemMsg, err := calculateSystemHealth(serviceStats)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate system health: %w", err)
	}

	// Calculating service health
	serviceScore, serviceMsg, err := calculateServiceHealth(serviceStats)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate service health: %w", err)
	}

	// Trigger alert webhooks on health score breach (< 70%)
	if systemScore < 70 {
		alerting.TriggerAlert(common.GetServiceName(), systemScore, "System Health Alert: "+systemMsg)
	}
	if serviceScore < 70 {
		alerting.TriggerAlert(common.GetServiceName(), serviceScore, "Service Health Alert: "+serviceMsg)
	}

	t := GetThresholds()
	return &models.SystemHealthInPercent{
		SystemHealth: models.HealthFields{
			Percentage:    common.RoundFloat64(systemScore, 2),
			AllowedByUser: t.MaxCPUUsage,
			Message:       systemMsg,
		},
		ServiceHealth: models.HealthFields{
			Percentage:    common.RoundFloat64(serviceScore, 2),
			AllowedByUser: t.MaxCPUUsage,
			Message:       serviceMsg,
		},
	}, nil
}
