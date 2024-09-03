package core

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

func GetCoreStatistics() models.CoreStatistics {
	rcount, durationTook := GetServiceMetrics()

	serviceInfo := common.GetServiceInfo()

	// 7.051466958s
	uptime := time.Since(serviceInfo.ServiceStartTime)
	uptimeStr := fmt.Sprintf("%.2f s", uptime.Seconds())

	if uptime.Seconds() > 60 {
		uptimeStr = fmt.Sprintf("%.2f m", uptime.Minutes())
	} else if uptime.Hours() > 60 {
		uptimeStr = fmt.Sprintf("%.2f h", uptime.Hours())
	} else if uptime.Hours() > 24 {
		uptimeStr = fmt.Sprintf("%.2f d", uptime.Hours()/24)
	} else if uptime.Hours() > 30*24 {
		uptimeStr = fmt.Sprintf("%.2f m", uptime.Hours()/(30*24))
	} else if uptime.Hours() > 12*30*24 {
		uptimeStr = fmt.Sprintf("%.2f y", uptime.Hours()/(12*30*24))
	}

	return models.CoreStatistics{
		Goroutines:                 runtime.NumGoroutine(),
		RequestCount:               rcount,
		Uptime:                     uptimeStr,
		TotalDurationTookByRequest: durationTook,
	}
}

func GetLoadStatistics() models.LoadStatistics {

	serviceCPU, systemCPU, totalCPU := common.GetCPULoad()

	// fmt.Printf("Service CPU Load: %.2f%%\n", serviceCPU)
	// fmt.Printf("System CPU Load: %.2f%%\n", systemCPU)
	// fmt.Printf("Total CPU Load: %.2f%%\n", totalCPU)

	serviceMem, systemMem, totalMem := common.GetMemoryLoad()
	// fmt.Printf("Service Memory Usage: %.2f%%\n", serviceMem)
	// fmt.Printf("System Memory Usage: %.2f%%\n", systemMem)
	// fmt.Printf("Total Memory Available: %.2f MB\n", totalMem/1024/1024)

	// serviceDisk, systemDisk, totalDisk := common.GetDiskLoad()
	// fmt.Printf("Service Disk Usage: %.2f MB\n", serviceDisk/1024/1024)
	// fmt.Printf("System Disk Usage: %.2f%%\n", systemDisk)
	// fmt.Printf("Total Disk Capacity: %.2f GB\n", totalDisk/1024/1024/1024)

	return models.LoadStatistics{
		ServiceCPULoad:       serviceCPU,
		SystemCPULoad:        systemCPU,
		TotalCPULoad:         totalCPU,
		ServiceMemLoad:       serviceMem,
		SystemMemLoad:        systemMem,
		TotalMemLoad:         common.ConvertToReadableUnit(totalMem),
		OverallLoadOfService: CalculateOverallLoad(serviceCPU, serviceMem),
		// ServiceDiskLoad: common.ParseFloat64ToString(serviceDisk),
		// SystemDiskLoad:  common.ParseFloat64ToString(systemDisk),
		// TotalDiskLoad:   common.ParseFloat64ToString(totalDisk),
	}
}

// Function to calculate overall load
func CalculateOverallLoad(serviceCPU, serviceMem string) string {

	// string to float64 conversion
	serviceCPUF := common.ParseStringToFloat64(serviceCPU)
	serviceMemF := common.ParseStringToFloat64(serviceMem)

	cpuWeight := 0.5 // Weight for CPU load
	memWeight := 0.5 // Weight for memory usage

	overallLoad := (cpuWeight * serviceCPUF) + (memWeight * serviceMemF) // Calculate overall load using weighted average

	if overallLoad > 100 {
		overallLoad = 100
	}

	return common.ParseFloat64ToString(overallLoad) + "%"
}

// CalculateHealthScore calculates a health score based on CPU and memory usage.
func CalculateHealthScore(serviceCPU, systemCPU, totalCPU float64, serviceMem, systemMem, totalMem float64) string {
	thresholds := GetServiceHealthThresholdsModel()

	// Calculate scores with bounds checks
	loadScore := (thresholds.MaxCPULoad.Value - totalCPU) / thresholds.MaxCPULoad.Value * thresholds.MaxCPULoad.Weight
	if loadScore < 0 {
		loadScore = 0
	}

	memoryScore := (thresholds.MaxMemory.Value - totalMem) / thresholds.MaxMemory.Value * thresholds.MaxMemory.Weight
	if memoryScore < 0 {
		memoryScore = 0
	}

	goroutineScore := (float64(thresholds.MaxGoroutines.Value) - float64(runtime.NumGoroutine())) / float64(thresholds.MaxGoroutines.Value) * thresholds.MaxGoroutines.Weight
	if goroutineScore < 0 {
		goroutineScore = 0
	}

	overallHealth := loadScore + memoryScore + goroutineScore

	// Ensure the health percent is within 0-100%
	if overallHealth > 100 {
		overallHealth = 100
	}

	return common.ParseFloat64ToString(overallHealth*100) + "%"
}

// In the CalculateHealthScore function, Weight is used to determine the importance of each metric (CPU load, memory usage, goroutines) in the overall health score.

// Weight Explanation
// Weight is a factor used to adjust the contribution of each metric to the overall health score. It represents the relative importance of the metric:

// Higher Weight: Indicates that the metric is more important in calculating the overall health score.
// Lower Weight: Indicates that the metric is less influential.

// Metric Score = ((Threshold Value - Actual Value) / Threshold Value) * Weight

// Usage:

// Load Score: Weighted by MaxLoad.Weight
// Memory Score: Weighted by MaxMemory.Weight
// Goroutine Score: Weighted by MaxGoroutines.Weight
// The combined weighted scores are summed to provide the final health score, ensuring that more critical metrics have a greater impact on the overall assessment.

// Example:
// If MaxLoad.Weight is 0.4, it means CPU load contributes 40% to the final health score.

// ### Health Scoring System

// The health score is calculated using weighted metrics to reflect their importance. Adjust the weights to prioritize metrics according to your application's needs.

// - **Weight**: Determines the relative importance of each metric in the overall health score.
//   - `MaxLoad.Weight`: Weight for CPU load. A higher weight indicates greater importance.
//   - `MaxMemory.Weight`: Weight for memory usage.
//   - `MaxGoroutines.Weight`: Weight for the number of goroutines.

// **Example**:
// - If CPU load is critical, set `MaxLoad.Weight` to 0.5.
// - For moderate importance, set `MaxMemory.Weight` to 0.3.
// - For less critical metrics, set `MaxGoroutines.Weight` to 0.2.

// Adjust these weights and thresholds based on your application's operational requirements.

// GetCPUStatistics retrieves the CPU statistics.
func GetCPUStatistics() models.CPUStatistics {
	var cpuStats models.CPUStatistics

	sysCPUPercent := GetCPUPrecent()
	memInfo := GetVirtualMemoryStats()

	procCPUPercent, _, err := getProcessUsage(common.GetProcessObject(), &memInfo)
	if err != nil {
		log.Panicf("Error fetching process usage: %v\n", err)
	}

	totalLogicalCores, _ := cpu.Counts(true)
	totalCores, _ := cpu.Counts(false)
	systemUsedCores := (sysCPUPercent / 100) * float64(totalLogicalCores)
	processUsedCores := (procCPUPercent / 100) * float64(totalLogicalCores)

	cpuStats.TotalCores = float64(totalCores)
	cpuStats.TotalLogicalCores = float64(totalLogicalCores)
	cpuStats.CoresUsedBySystem = common.RoundFloat64(systemUsedCores, 3)
	cpuStats.CoresUsedByService = common.RoundFloat64(processUsedCores, 3)

	// Converting CPU usage to percentage strings
	cpuStats.CoresUsedBySystemInPercent = strconv.FormatFloat(cpuStats.CoresUsedBySystem, 'f', 2, 64) + "%"
	cpuStats.CoresUsedByServiceInPercent = strconv.FormatFloat(cpuStats.CoresUsedByService, 'f', 2, 64) + "%"

	return cpuStats
}

// GetMemoryStatistics retrieves memory statistics.
func GetMemoryStatistics() models.MemoryStatistics {

	memInfo, err := mem.VirtualMemory() // Fetcing system memory statistics
	if err != nil {
		log.Fatalf("Error fetching virtual memory info: %v", err)
	}

	swapInfo, err := mem.SwapMemory() // Fetching swap memory statistics
	if err != nil {
		log.Fatalf("Error fetching swap memory info: %v", err)
	}

	m := ReadMemStats() // Get the memory statistics for the service
	return models.MemoryStatistics{
		TotalSystemMemory:   common.BytesToUnit(memInfo.Total),
		MemoryUsedBySystem:  common.BytesToUnit(memInfo.Used),
		AvailableMemory:     common.BytesToUnit(memInfo.Available),
		TotalSwapMemory:     common.BytesToUnit(swapInfo.Total),
		FreeSwapMemory:      common.BytesToUnit(swapInfo.Free),
		MemoryUsedByService: common.BytesToUnit(m.Alloc), // Example metric
		StackMemoryUsage:    common.BytesToUnit(m.StackInuse),
		GCPauseDuration:     fmt.Sprintf("%.2f ms", float64(m.PauseTotalNs)/float64(time.Millisecond)), // Convert nanoseconds to milliseconds
		MemStatsRecords:     ConstructMemStats(m),
	}
}

// ConstructMemStats constructs a list of memory statistics records.
func ConstructMemStats(memStats *runtime.MemStats) []models.Record {
	r := []models.Record{
		newRecord("Alloc", "Bytes of allocated heap objects.", memStats.Alloc),
		newRecord("TotalAlloc", "Cumulative bytes allocated for heap objects.", memStats.TotalAlloc),
		newRecord("Sys", "Total bytes of memory obtained from the OS.", memStats.Sys),
		newRecord("Lookups", "Number of pointer lookups performed by the runtime.", memStats.Lookups),
		newRecord("Mallocs", "Cumulative count of heap objects allocated.", memStats.Mallocs),
		newRecord("Frees", "Cumulative count of heap objects freed.", memStats.Frees),
		newRecord("HeapAlloc", "Bytes of allocated heap objects.", memStats.HeapAlloc),
		newRecord("HeapSys", "Bytes of heap memory obtained from the OS.", memStats.HeapSys),
		newRecord("HeapIdle", "Bytes in idle (unused) spans.", memStats.HeapIdle),
		newRecord("HeapInuse", "Bytes in in-use spans.", memStats.HeapInuse),
		newRecord("HeapReleased", "Bytes of physical memory returned to the OS.", memStats.HeapReleased),
		newRecord("HeapObjects", "Number of allocated heap objects.", memStats.HeapObjects),
		newRecord("StackInuse", "Bytes in stack spans.", memStats.StackInuse),
		newRecord("StackSys", "Bytes of stack memory obtained from the OS.", memStats.StackSys),
		newRecord("MSpanInuse", "Bytes of allocated mspan structures.", memStats.MSpanInuse),
		newRecord("MSpanSys", "Bytes of memory obtained from the OS for mspan structures.", memStats.MSpanSys),
		newRecord("MCacheInuse", "Bytes of allocated mcache structures.", memStats.MCacheInuse),
		newRecord("MCacheSys", "Bytes of memory obtained from the OS for mcache structures.", memStats.MCacheSys),
		newRecord("BuckHashSys", "Bytes of memory in profiling bucket hash tables.", memStats.BuckHashSys),
		newRecord("GCSys", "Bytes of memory in garbage collection metadata.", memStats.GCSys),
		newRecord("OtherSys", "Bytes of memory in miscellaneous off-heap runtime allocations.", memStats.OtherSys),
		newRecord("NextGC", "Target heap size of the next GC cycle.", memStats.NextGC),
		newRecord("LastGC", "Time the last garbage collection finished (nanoseconds since the UNIX epoch).", memStats.LastGC),
		newRecord("PauseTotalNs", "Cumulative nanoseconds in GC stop-the-world pauses since program start.", memStats.PauseTotalNs),
		newRecord("NumGC", "Number of completed GC cycles.", uint64(memStats.NumGC)),
		newRecord("NumForcedGC", "Number of GC cycles that were forced by the application calling GC.", uint64(memStats.NumForcedGC)),
		newRecord("GCCPUFraction", "Fraction of this program's available CPU time used by the GC.", memStats.GCCPUFraction),
	}

	return r
}

// func GetCPUUsedInPercentByService() float64 {
// 	_, proc := common.GetProcessDetails()
// 	procCPUPercent, _, err := getProcessUsage(proc, nil)
// 	if err != nil {
// 		log.Fatalf("Error fetching process usage: %v\n", err)
// 	}
// 	return procCPUPercent
// }

// GetNetworkIO retrieves network I/O statistics.
func GetNetworkIO() (float64, float64) {
	// Fetch network I/O statistics
	netIO, err := net.IOCounters(true)
	if err != nil {
		log.Fatalf("Error fetching network I/O statistics: %v", err)
	}

	var totalBytesReceived, totalBytesSent float64

	// Aggregate statistics from all network interfaces
	for _, iface := range netIO {
		totalBytesReceived += float64(iface.BytesRecv)
		totalBytesSent += float64(iface.BytesSent)
	}

	return totalBytesReceived, totalBytesSent
}

// GetServiceMetrics retrieves the service metrics.
func GetServiceHealth(loadStats *models.LoadStatistics) models.ServiceHealth {

	overallHealth := CalculateHealthScore(
		common.ParseStringToFloat64(loadStats.ServiceCPULoad),
		common.ParseStringToFloat64(loadStats.SystemCPULoad),
		common.ParseStringToFloat64(loadStats.TotalCPULoad),
		common.ParseStringToFloat64(loadStats.ServiceMemLoad),
		common.ParseStringToFloat64(loadStats.SystemMemLoad),
		common.ParseStringToFloat64(loadStats.TotalMemLoad)*1024*1024,
	)

	healthy := true
	message := ""

	if overallHealth > "50%" {
		message = "[Healthy] Service is healthy and running smoothly."
	} else {
		healthy = false
		message = "[Unhealthy] Service is under heavy load or high memory usage, consider scaling up or optimizing. Check the memory and CPU usage for more details."
	}

	return models.ServiceHealth{
		OverallHealthPercent: overallHealth,
		Health: models.Health{
			Healthy: healthy,
			Message: message,
		},
	}
}
