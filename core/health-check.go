package core

import (
	"runtime"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
)

// getProcessCPUUsage returns the CPU usage of the process
func getProcessCPUUsage() (float64, error) {
	return common.GetProcessObject().CPUPercent()
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

// calculateHealthScore calculates the health score based on usage and thresholds
func calculateHealthScore(usage, maxThreshold float64) float64 {
	return (usage / maxThreshold) * 100
}

// calculateServiceHealth calculates service health based on CPU, memory, and goroutines
func calculateServiceHealth(stats *models.ServiceStats) (float64, error) {
	cpuUsage, err := getProcessCPUUsage() // Get CPU usage
	if err != nil {
		return 0, err
	}

	memoryUsagePercentage, err := calculateMemoryUsagePercentage(stats.MemoryStatistics.MemoryUsedByService, stats.MemoryStatistics.TotalSystemMemory)
	if err != nil {
		return 0, err
	}

	// Calculating health scores
	cpuHealthScore := calculateHealthScore(cpuUsage, serviceHealthThresholds.MaxCPUUsage)
	memoryHealthScore := calculateHealthScore(memoryUsagePercentage, serviceHealthThresholds.MaxMemoryUsage)
	goroutinesHealthScore := calculateHealthScore(float64(runtime.NumGoroutine()), float64(serviceHealthThresholds.MaxGoRoutines))

	// Final health score calculation
	finalScore := 100 - ((cpuHealthScore + memoryHealthScore + goroutinesHealthScore) / 3)
	return finalScore, nil
}

// calculateSystemHealth calculates system health based on CPU and memory
func calculateSystemHealth(stats *models.ServiceStats) (float64, error) {
	memoryUsagePercentage, err := calculateMemoryUsagePercentage(stats.MemoryStatistics.MemoryUsedBySystem, stats.MemoryStatistics.TotalSystemMemory)
	if err != nil {
		return 0, err
	}

	sysCPUUsage, err := getProcessCPUUsage()
	if err != nil {
		return 0, err
	}

	// Calculating the health scores
	cpuHealthScore := calculateHealthScore(sysCPUUsage, serviceHealthThresholds.MaxCPUUsage)
	memoryHealthScore := calculateHealthScore(memoryUsagePercentage, serviceHealthThresholds.MaxMemoryUsage)

	// Final system health score calculation
	finalScore := 100 - ((cpuHealthScore + memoryHealthScore) / 2)
	return finalScore, nil
}

// CalculateHealthScore calculates the health score of the system and the service
func CalculateHealthScore(serviceStats *models.ServiceStats) (*models.SystemHealthInPercent, error) {
	sysScore, err := calculateSystemHealth(serviceStats)
	if err != nil {
		return nil, err
	}

	servScore, err := calculateServiceHealth(serviceStats)
	if err != nil {
		return nil, err
	}

	overallHealth := (sysScore + servScore) / 2
	return &models.SystemHealthInPercent{
		SystemHealth:  common.RoundFloat64(sysScore, 2),
		ServiceHealth: common.RoundFloat64(servScore, 2),
		OverallHealth: common.RoundFloat64(overallHealth, 0),
	}, nil
}
