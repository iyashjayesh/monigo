package core

import (
	"runtime"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
	"github.com/shirou/gopsutil/cpu"
)

func getSystemCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}
	return percentages[0], nil
}

func getSystemMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	totalMemory := float64(m.Sys) / (1024 * 1024)
	usedMemory := float64(m.Alloc) / (1024 * 1024)
	return (usedMemory / totalMemory) * 100
}

func getGoroutineCount() int {
	return runtime.NumGoroutine()
}

func getProcessCPUUsage() (float64, error) {
	percent, err := common.GetProcessObject().CPUPercent()
	if err != nil {
		return 0, err
	}
	return percent, nil
}

func getProcessMemoryUsage() (float64, error) {
	memInfo, err := common.GetProcessObject().MemoryInfo()
	if err != nil {
		return 0, err
	}
	totalMemory := float64(memInfo.RSS) / (1024 * 1024)
	return totalMemory, nil
}

func getHealthScore(usage float64, thresholds models.ServiceHealthThresholds) int {
	switch {
	case usage <= thresholds.Low:
		return 10
	case usage <= thresholds.Medium:
		return 7
	case usage <= thresholds.High:
		return 4
	default:
		return 1
	}
}

func getGoroutineHealthScore(goroutines int, thresholds models.ServiceHealthThresholds) int {
	switch {
	case goroutines <= thresholds.GoRoutinesLow:
		return 10
	case goroutines <= thresholds.GoRoutinesHigh:
		return 7
	default:
		return 4
	}
}

func calculateServiceHealth(thresholds models.ServiceHealthThresholds) (float64, error) {
	cpuUsage, err := getProcessCPUUsage()
	if err != nil {
		return 0, err
	}

	memoryUsage, err := getProcessMemoryUsage()
	if err != nil {
		return 0, err
	}

	cpuScore := getHealthScore(cpuUsage, thresholds)
	memoryScore := getHealthScore(memoryUsage, thresholds)

	healthScore := float64(cpuScore+memoryScore) / 2
	healthScore = healthScore * 10

	return healthScore, nil
}

func calculateSystemHealth(thresholds models.ServiceHealthThresholds) (float64, error) {
	systemCPUUsage, err := getSystemCPUUsage()
	if err != nil {
		return 0, err
	}

	systemMemoryUsage := getSystemMemoryUsage()

	serviceCPUUsage, err := getProcessCPUUsage()
	if err != nil {
		return 0, err
	}

	serviceMemoryUsage, err := getProcessMemoryUsage()
	if err != nil {
		return 0, err
	}

	adjustedCPUUsage := systemCPUUsage - serviceCPUUsage
	adjustedMemoryUsage := systemMemoryUsage - (serviceMemoryUsage * 100 / (1024 * 1024))

	if adjustedCPUUsage < 0 {
		adjustedCPUUsage = 0
	}
	if adjustedMemoryUsage < 0 {
		adjustedMemoryUsage = 0
	}

	goroutines := getGoroutineCount()

	cpuScore := getHealthScore(adjustedCPUUsage, thresholds)
	memoryScore := getHealthScore(adjustedMemoryUsage, thresholds)
	goroScore := getGoroutineHealthScore(goroutines, thresholds)

	healthScore := float64(cpuScore+memoryScore+goroScore) / 3
	healthScore = healthScore * 10

	return healthScore, nil
}

func CalculateHealthScore(thresholds models.ServiceHealthThresholds) (*models.SystemHealthInPercent, error) {
	systemHealth, err := calculateSystemHealth(thresholds)
	if err != nil {
		return nil, err
	}

	serviceHealth, err := calculateServiceHealth(thresholds)
	if err != nil {
		return nil, err
	}

	overallHealth := (systemHealth + serviceHealth) / 2

	return &models.SystemHealthInPercent{
		SystemHealth:  systemHealth,
		ServiceHealth: serviceHealth,
		OverallHealth: overallHealth,
	}, nil
}
