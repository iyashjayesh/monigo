package core

import (
	"runtime"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/iyashjayesh/monigo/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var (
	mu                      sync.Mutex
	serviceHealthThresholds models.ServiceHealthThresholds
)

// GetCPUPrecent returns the total number of requests
func GetCPUPrecent() (float64, error) {
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		logger.Log.Error("Error fetching CPU usage", "error", err)
		return 0, err
	}

	var total float64
	for _, percent := range cpuPercents {
		total += percent
	}

	return total / float64(len(cpuPercents)), nil
}

// GetVirtualMemoryStats returns the virtual memory statistics
func GetVirtualMemoryStats() (mem.VirtualMemoryStat, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("Error fetching memory usage", "error", err)
		return mem.VirtualMemoryStat{}, err
	}

	return *memInfo, nil
}

// Fetches and returns process CPU and memory usage
func getProcessUsage(proc *process.Process, memsStats *mem.VirtualMemoryStat) (float64, float64, error) {
	procCPUPercent, err := proc.CPUPercent()
	if err != nil {
		return 0, 0, err
	}

	memStats := ReadMemStats()

	// Calculate memory used by the process as a percentage of total system memory
	processMemPercent := (float64(memStats.Alloc) / float64(memsStats.Total)) * 100

	return procCPUPercent, processMemPercent, nil
}

// SetServiceThresholds sets the service thresholds to calculate the overall service health.
func ConfigureServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) {
	serviceHealthThresholds = *thresholdsValues
}

// newRecord creates a new Record with appropriate units and human-readable formats.
func newRecord(name, description string, value interface{}) models.Record {
	switch v := value.(type) {
	case uint64:
		size, unit := common.ConvertToReadableSize(v)
		return models.Record{
			Name:        name,
			Description: description,
			Value:       size,
			Unit:        unit,
		}
	case float64:
		return models.Record{
			Name:        name,
			Description: description,
			Value:       v,
			Unit:        "fraction",
		}
	default:
		return models.Record{
			Name:        name,
			Description: description,
			Value:       value,
		}
	}
}

// ReadMemStats reads and returns the memory statistics
func ReadMemStats() *runtime.MemStats {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	return &memStats
}
