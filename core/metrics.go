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
	functionMetrics         = make(map[string]*models.FunctionMetrics)
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

func GetFunctionMetrics(functionName string) *models.FunctionMetrics {
	mu.Lock()
	defer mu.Unlock()
	return functionMetrics[functionName]
}

func GetProcessSats() models.ProcessStats {

	pid, proc := common.GetProcessDetails()
	sysCPUPercent := GetCPUPrecent()
	memInfo := GetVirtualMemoryStats()

	procCPUPercent, procMemPercent, err := getProcessUsage(proc, &memInfo)
	if err != nil {
		log.Panicf("Error fetching process usage: %v\n", err)
		return models.ProcessStats{}
	}

	totalCores, _ := cpu.Counts(false)
	totalLogicalCores := runtime.NumCPU()
	systemUsedCores := (sysCPUPercent / 100) * float64(totalLogicalCores)
	processUsedCores := (procCPUPercent / 100) * float64(totalLogicalCores)

	return models.ProcessStats{
		ProcessId:      pid,
		SysCPUPercent:  sysCPUPercent,
		ProcCPUPercent: procCPUPercent,
		ProcMemPercent: procMemPercent,
		// TotalMemoryUsage:  sysMemUsage.TotalMemoryUsage,
		// FreeMemory:        sysMemUsage.FreeMemory,
		// UsedMemoryPercent: sysMemUsage.UsedPercent,
		TotalCores:        totalCores,
		TotalLogicalCores: totalLogicalCores,
		SystemUsedCores:   systemUsedCores,
		ProcessUsedCores:  processUsedCores,
	}
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

func GetLocalFunctionMetrics() map[string]*models.FunctionMetrics {
	return functionMetrics
}

// func GetServiceMetricsModel() models.ServiceMetrics {

// 	requestCount, totalDuration := GetServiceMetrics()
// 	serviceStat := GetProcessSats()
// 	memStats := ReadMemStats()

// 	// SystemUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.SystemUsedCores)
// 	// ProcessUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.ProcessUsedCores)

// 	// cores := ProcessUsedCoresToString + "PC / " +
// 	// 	SystemUsedCoresToString + "SC / " +
// 	// 	strconv.Itoa(serviceStat.TotalLogicalCores) + "LC / " +
// 	// 	strconv.Itoa(serviceStat.TotalCores) + "C"

// 	metrics := models.ServiceMetrics{
// 		Load:                   serviceStat.ProcCPUPercent,
// 		Cores:                  serviceStat.ProcessUsedCores,
// 		MemoryUsed:             float64(memStats.Alloc),
// 		NumberOfReqServerd:     float64(requestCount),
// 		GoRoutines:             float64(runtime.NumGoroutine()),
// 		TotalAlloc:             float64(memStats.TotalAlloc),
// 		MemoryAllocSys:         float64(memStats.Sys),
// 		HeapAlloc:              float64(memStats.HeapAlloc),
// 		HeapAllocSys:           float64(memStats.HeapSys),
// 		TotalDurationTookByAPI: totalDuration,
// 	}

// 	return metrics
// }

func CalculateServiceHealth(metrics models.ServiceMetrics) models.ServiceHealth {
	// goroutines := strconv.Itoa(int(metrics.GoRoutines))
	// requests := strconv.Itoa(int(metrics.NumberOfReqServerd))
	memoryUsed := metrics.MemoryUsed
	cpuLoad := metrics.Load

	health := "Healthy"
	healthy := true
	if memoryUsed > 80.0 {
		health = "Warning: High Memory Usage"
		healthy = false
	}
	if memoryUsed > 80.0 || cpuLoad > 80.0 {
		health = "Critical: Service Under Heavy Load"
		healthy = false
	}

	// OverallHealthPercent
	// overallHealthPercent := CalculateOverallHealth(&metrics)

	// strToInt := func(s string) int {
	// 	i, _ := strconv.Atoi(s)
	// 	return i
	// }

	return models.ServiceHealth{
		// Goroutines:           strToInt(goroutines),
		// Requests:             strToInt(requests),
		// MemoryUsed:           memoryUsed,
		// CPUPercent:           cpuLoad,
		// OverallHealthPercent: overallHealthPercent,
		Health: models.Health{
			Healthy: healthy,
			Message: health,
		},
	}
}

// Example calculation in Go
func CalculateOverallHealth(metrics *models.ServiceMetrics) float64 {

	// Calculating the health score for each metric with a weight
	loadScore := (serviceHealthThresholds.MaxCPULoad.Value - metrics.Load) / serviceHealthThresholds.MaxCPULoad.Value * serviceHealthThresholds.MaxCPULoad.Weight                                                // 25% weight
	memoryScore := (serviceHealthThresholds.MaxMemory.Value - metrics.MemoryUsed) / serviceHealthThresholds.MaxMemory.Value * serviceHealthThresholds.MaxMemory.Weight                                           // 25% weight
	goroutineScore := (float64(serviceHealthThresholds.MaxGoroutines.Value) - float64(metrics.GoRoutines)) / float64(serviceHealthThresholds.MaxGoroutines.Value) * serviceHealthThresholds.MaxGoroutines.Weight // 20% weight
	// requestScore := (float64(serviceHealthThresholds.MaxRequests.Value) - float64(requests)) / float64(serviceHealthThresholds.MaxRequests.Value) * serviceHealthThresholds.MaxRequests.Weight                   // 15% weight
	// uptimeScore := (uptime / 1440.0) * 15                                                                                                                                                                        // 15% weight, normalized for uptime (example 1440 mins = 1 day)

	// Combine into an overall health percentage
	overallHealth := loadScore + memoryScore + goroutineScore
	// + requestScore + uptimeScore

	// Ensure the health percent is within 0-100%
	if overallHealth > 100 {
		overallHealth = 100
	} else if overallHealth < 0 {
		overallHealth = 0
	}

	return overallHealth
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

func GetSystemCPUInfo() models.CPUStat {
	// numCPU := float64(runtime.NumCPU())
	serviceStat := GetProcessSats()

	// processUsedCoresInPercent := (float64(serviceStat.ProcessUsedCores) / float64(serviceStat.TotalLogicalCores)) * 100

	return models.CPUStat{
		TotalCores:        float64(serviceStat.TotalCores),
		TotalLogicalCores: float64(serviceStat.TotalLogicalCores),
		// SystemUsedCores:   serviceStat.SystemUsedCores,
		// ProcessUsedCores:  serviceStat.ProcessUsedCores,
		// Cores:             fmt.Sprintf("%.0f", numCPU),
		// UsedInPercent:     fmt.Sprintf("%.2f%%", processUsedCoresInPercent),
	}
}

func GetSystemMemoryInfo() models.MemoryStat {

	// vm := GetVirtualMemoryStats() // Get the virtual memory statistics
	// memStats := ReadMemStats() // Get the memory statistics

	return models.MemoryStat{
		// TotalMemory:         float64(vm.Total),
		// UsedBySystem:        float64(vm.Used),
		// FreeMemory:          float64(vm.Free),
		// UsedByProcess:       float64(memStats.Alloc),
		// HeapAllocByProcess:  float64(memStats.HeapAlloc),
		// HeapSysByProcess:    float64(memStats.HeapSys),
		// TotalAllocByProcess: float64(memStats.TotalAlloc),
		// TotalSysByProcess:   float64(memStats.Sys),
		// UsedInPercent:   fmt.Sprintf("%.2f%%", (float64(memStats.Alloc)/float64(vm.Total))*100),
		// MemStatsRecords: ConstructMemStats(memStats),
	}
}

func ReadMemStats() *runtime.MemStats {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	return &memStats
}
