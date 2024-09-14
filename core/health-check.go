package core

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"

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
	log.Println("memInfo ", memInfo)
	log.Println("memInfo.RSS ", memInfo.RSS)
	usedMemory := float64(memInfo.RSS) / (1024 * 1024)

	// totalaVailableMemory, err := common.GetTotalMemory()
	// if err != nil {
	// 	return 0, err
	// }

	return usedMemory, nil // return in MB
}

// func getHealthScore(usage float64, thresholds models.ServiceHealthThresholds) int {
// 	switch {
// 	case usage <= thresholds.Low:
// 		return 10
// 	case usage <= thresholds.Medium:
// 		return 7
// 	case usage <= thresholds.High:
// 		return 4
// 	default:
// 		return 1
// 	}
// }

// func getGoroutineHealthScore(goroutines int, thresholds models.ServiceHealthThresholds) int {
// 	switch {
// 	case goroutines <= thresholds.GoRoutinesLow:
// 		return 10
// 	case goroutines <= thresholds.GoRoutinesHigh:
// 		return 7
// 	default:
// 		return 4
// 	}
// }

// Helper function to convert values to MB
func convertToMB(value string) (float64, error) {
	value = strings.TrimSpace(value)
	value = strings.Replace(value, " ", "", -1)
	unit := strings.ToUpper(value[len(value)-2:])
	val, err := strconv.ParseFloat(value[:len(value)-2], 64)
	if err != nil {
		return 0, err
	}

	unit = strings.ToUpper(unit)
	switch unit {
	case "TB":
		return val * 1024 * 1024, nil
	case "GB":
		return val * 1024, nil
	case "MB":
		return val, nil
	case "KB":
		return val / 1024, nil
	default:
		return 0, fmt.Errorf("unsupported memory unit: %s", unit)
	}
}

// Calculate the service health based on CPU and memory usage
func calculateServiceHealth(stats *models.ServiceStats) (float64, error) {

	cpuUsage, err := getProcessCPUUsage() // Getting the CPU usage
	if err != nil {
		return 0, err
	}

	totalMemoryMB, err := convertToMB(stats.MemoryStatistics.TotalSystemMemory)
	if err != nil {
		return 0, err
	}
	usedMemoryMB, err := convertToMB(stats.MemoryStatistics.MemoryUsedByService)
	if err != nil {
		return 0, err
	}

	// Calculating memory usage percentage
	memoryUsagePercentage := (usedMemoryMB / totalMemoryMB) * 100
	goroutines := getGoroutineCount()

	// Checking if service is within health thresholds
	// cpuHealthy := cpuUsage < serviceHealthThresholds.MaxCPUUsage
	// memoryHealthy := memoryUsagePercentage < serviceHealthThresholds.MaxMemoryUsage
	// goroutinesHealthy := goroutines < serviceHealthThresholds.MaxGoRoutines

	// health score
	cpuHealthScore := (cpuUsage / serviceHealthThresholds.MaxCPUUsage) * 100
	memoryHealthScore := (memoryUsagePercentage / serviceHealthThresholds.MaxMemoryUsage) * 100
	goroutinesHealthScore := (float64(goroutines) / float64(serviceHealthThresholds.MaxGoRoutines)) * 100
	finalScore := 100 - ((cpuHealthScore + memoryHealthScore + goroutinesHealthScore) / 3)

	return finalScore, nil
}

// Calculate the system health based on CPU and memory usage
func calculateSystemHealth(stats *models.ServiceStats) (float64, error) {

	totalMemoryMB, err := convertToMB(stats.MemoryStatistics.TotalSystemMemory)
	if err != nil {
		return 0, err
	}

	memotryUsedBySystem, err := convertToMB(stats.MemoryStatistics.MemoryUsedBySystem)
	if err != nil {
		return 0, err
	}

	totalUsedBySystem := (memotryUsedBySystem / totalMemoryMB) * 100
	sysCPUUsage, err := getProcessCPUUsage()
	if err != nil {
		return 0, err
	}

	// cpuHealthy := sysCPUUsage < serviceHealthThresholds.MaxCPUUsage
	// memoryHealthy := totalUsedBySystem < serviceHealthThresholds.MaxMemoryUsage

	cpuHealthScore := (sysCPUUsage / serviceHealthThresholds.MaxCPUUsage) * 100
	memoryHealthScore := (totalUsedBySystem / serviceHealthThresholds.MaxMemoryUsage) * 100
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
		SystemHealth:  sysScore,
		ServiceHealth: servScore,
		OverallHealth: overallHealth,
	}, nil
}
