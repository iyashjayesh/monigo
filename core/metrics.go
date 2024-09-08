package core

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var (
	mu                      sync.Mutex
	requestCount            int64
	totalDuration           time.Duration
	serviceHealthThresholds = models.ServiceHealthThresholds{ // Default thresholds
		MaxGoroutines: models.Thresholds{
			Value:  100,
			Weight: 0.2,
		},
		MaxCPULoad: models.Thresholds{
			Value:  85,
			Weight: 0.7,
		},
		MaxMemory: models.Thresholds{
			Value:  85,
			Weight: 0.7,
		},
	}
)

func RecordRequestDuration(duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	requestCount++
	totalDuration += duration
}

func GetServiceMetrics() (int64, time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	return requestCount, totalDuration
}

func GetCPUPrecent() float64 {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Panicf("Error fetching CPU usage: %v\n", err)
		return 0
	}
	return cpuPercent[0]
}

func GetVirtualMemoryStats() mem.VirtualMemoryStat {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Panicf("Error fetching memory usage: %v\n", err)
		return mem.VirtualMemoryStat{}
	}

	return *memInfo
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
func SetServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) {
	serviceHealthThresholds = *thresholdsValues
}

func GetServiceHealthThresholdsModel() models.ServiceHealthThresholds {
	return serviceHealthThresholds
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

func ReadMemStats() *runtime.MemStats {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	return &memStats
}
