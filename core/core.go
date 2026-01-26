package core

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

// GetServiceStats collects statistics related to service and system performance.
func GetServiceStats() models.ServiceStats {

	var stats models.ServiceStats
	stats.CoreStatistics = GetCoreStatistics()

	var wg sync.WaitGroup
	wg.Add(6)

	// Goroutine to fetch load statistics
	go func() {
		defer wg.Done()
		stats.LoadStatistics = GetLoadStatistics()
	}()

	// Goroutine to fetch memory statistics
	go func() {
		defer wg.Done()
		stats.MemoryStatistics = GetMemoryStatistics()
	}()

	// Goroutine to fetch CPU statistics
	go func() {
		defer wg.Done()
		stats.CPUStatistics = GetCPUStatistics()
	}()

	// Goroutine to fetch memory allocation statistics
	go func() {
		defer wg.Done()
		memStats := ReadMemStats()
		stats.HeapAllocByService = common.BytesToUnit(memStats.HeapAlloc)
		stats.HeapAllocBySystem = common.BytesToUnit(memStats.HeapSys)
		stats.TotalAllocByService = common.BytesToUnit(memStats.TotalAlloc)
		stats.TotalMemoryByOS = common.BytesToUnit(memStats.Sys)
		stats.HeapAllocByServiceRaw = memStats.HeapAlloc
		stats.HeapAllocBySystemRaw = memStats.HeapSys
		stats.TotalAllocByServiceRaw = memStats.TotalAlloc
		stats.TotalMemoryByOSRaw = memStats.Sys
	}()

	// Goroutine to fetch network I/O statistics
	go func() {
		defer wg.Done()
		stats.NetworkIO.BytesReceived, stats.NetworkIO.BytesSent = GetNetworkIO()
	}()

	// Goroutine to fetch disk I/O statistics
	go func() {
		defer wg.Done()
		stats.DiskIO.ReadBytes, stats.DiskIO.WriteBytes = GetDiskIO()
	}()

	wg.Wait()

	stats.Health = GetServiceHealth(&stats)

	return stats
}

// formatUptime returns a formatted string based on the service uptime duration
func formatUptime(uptime time.Duration) string {
	hours := uptime.Hours()

	switch {
	case hours > 12*30*24: // More than a year
		return fmt.Sprintf("%.2f y", hours/(12*30*24))
	case hours > 30*24: // More than a month
		return fmt.Sprintf("%.2f mo", hours/(30*24))
	case hours > 24: // More than a day
		return fmt.Sprintf("%.2f d", hours/24)
	case hours > 1: // More than an hour
		return fmt.Sprintf("%.2f h", hours)
	case uptime.Minutes() > 1: // More than a minute
		return fmt.Sprintf("%.2f m", uptime.Minutes())
	default: // Less than a minute
		return fmt.Sprintf("%.2f s", uptime.Seconds())
	}
}

// GetCoreStatistics retrieves core statistics like goroutines, request count, uptime, and total request duration
func GetCoreStatistics() models.CoreStatistics {

	serviceInfo := common.GetServiceInfo()
	uptime := time.Since(serviceInfo.ServiceStartTime)
	uptimeFormatted := formatUptime(uptime)

	return models.CoreStatistics{
		Goroutines: runtime.NumGoroutine(),
		Uptime:     uptimeFormatted,
	}
}

// GetLoadStatistics retrieves load statistics for CPU, memory, and optionally disk usage.
func GetLoadStatistics() models.LoadStatistics {

	// Fetch CPU load statistics
	serviceCPULoad, systemCPULoad, totalCPULoad, serviceCPUF, systemCPUF, _ := common.GetCPULoad()

	// Fetch memory load statistics
	serviceMemLoad, systemMemLoad, totalMemAvailable, serviceMemF, systemMemF, _ := common.GetMemoryLoad()

	// Fetch disk load statistics
	serviceDisk, systemDisk, totalDisk, systemDiskF, totalDiskF := common.GetDiskLoad()

	overallLoadF, overallLoadStr := CalculateOverallLoad(serviceCPUF, serviceMemF)

	return models.LoadStatistics{
		ServiceCPULoad:          serviceCPULoad,
		SystemCPULoad:           systemCPULoad,
		TotalCPULoad:            totalCPULoad,
		ServiceMemLoad:          serviceMemLoad,
		SystemMemLoad:           systemMemLoad,
		TotalMemLoad:            totalMemAvailable,
		OverallLoadOfService:    overallLoadStr,
		ServiceDiskLoad:         serviceDisk,
		SystemDiskLoad:          systemDisk,
		TotalDiskLoad:           totalDisk,
		ServiceCPULoadRaw:       serviceCPUF,
		SystemCPULoadRaw:        systemCPUF,
		ServiceMemLoadRaw:       serviceMemF,
		SystemMemLoadRaw:        systemMemF,
		OverallLoadOfServiceRaw: overallLoadF,
		SystemDiskLoadRaw:       systemDiskF,
		TotalDiskLoadRaw:        totalDiskF,
	}
}

// Function to calculate overall load
func CalculateOverallLoad(serviceCPUF, serviceMemF float64) (float64, string) {
	cpuWeight := 0.5 // Weight for CPU load
	memWeight := 0.5 // Weight for memory usage

	overallLoad := (cpuWeight * serviceCPUF) + (memWeight * serviceMemF) // Calculate overall load using weighted average

	if overallLoad > 100 {
		overallLoad = 100
	}

	return overallLoad, common.ParseFloat64ToString(overallLoad) + "%"
}

// GetCPUStatistics retrieves the CPU statistics.
func GetCPUStatistics() models.CPUStatistics {
	var cpuStats models.CPUStatistics

	sysCPUPercent, err := GetCPUPrecent()
	if err != nil {
		log.Printf("[MoniGo] Error fetching system CPU percent: %v", err)
		sysCPUPercent = 0
	}
	memInfo, err := GetVirtualMemoryStats()
	if err != nil {
		log.Printf("[MoniGo] Error fetching virtual memory stats: %v", err)
		memInfo = mem.VirtualMemoryStat{}
	}

	procCPUPercent, _, err := getProcessUsage(common.GetProcessObject(), &memInfo)
	if err != nil {
		log.Printf("[MoniGo] Error fetching process usage: %v\n", err)
		procCPUPercent = 0
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
		log.Printf("[MoniGo] Error fetching virtual memory info: %v", err)
		return models.MemoryStatistics{}
	}

	swapInfo, err := mem.SwapMemory() // Fetching swap memory statistics
	if err != nil {
		log.Printf("[MoniGo] Error fetching swap memory info: %v", err)
		// valid SwapMemory struct to prevent nil pointer later if used, or continue with zeroed swapInfo
		swapInfo = &mem.SwapMemoryStat{}
	}

	m := ReadMemStats() // Get the memory statistics for the service
	return models.MemoryStatistics{
		TotalSystemMemory:      common.BytesToUnit(memInfo.Total),
		MemoryUsedBySystem:     common.BytesToUnit(memInfo.Used),
		AvailableMemory:        common.BytesToUnit(memInfo.Available),
		TotalSwapMemory:        common.BytesToUnit(swapInfo.Total),
		FreeSwapMemory:         common.BytesToUnit(swapInfo.Free),
		MemoryUsedByService:    common.BytesToUnit(m.Alloc), // Example metric
		StackMemoryUsage:       common.BytesToUnit(m.StackInuse),
		GCPauseDuration:        fmt.Sprintf("%.2f ms", float64(m.PauseTotalNs)/float64(time.Millisecond)), // Convert nanoseconds to milliseconds
		MemStatsRecords:        ConstructMemStats(m),
		RawMemStatsRecords:     ConstructRawMemStats(m),
		TotalSystemMemoryRaw:   float64(memInfo.Total),
		MemoryUsedBySystemRaw:  float64(memInfo.Used),
		MemoryUsedByServiceRaw: float64(m.Alloc),
		AvailableMemoryRaw:     float64(memInfo.Available),
		GCPauseDurationRaw:     float64(m.PauseTotalNs) / float64(time.Millisecond),
		StackMemoryUsageRaw:    float64(m.StackInuse),
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

// GetNetworkIO retrieves network I/O statistics.
func GetNetworkIO() (float64, float64) {
	// Fetch network I/O statistics
	netIO, err := net.IOCounters(true)
	if err != nil {
		log.Printf("[MoniGo] Error fetching network I/O statistics: %v", err)
		return 0, 0
	}

	var totalBytesReceived, totalBytesSent float64
	for _, iface := range netIO { // Aggregate statistics from all network interfaces
		totalBytesReceived += float64(iface.BytesRecv)
		totalBytesSent += float64(iface.BytesSent)
	}

	return totalBytesReceived, totalBytesSent
}

// getStatusMessage returns a status message based on the health score.
func getStatusMessage(healthScore float64) string {

	var message string
	switch {
	case healthScore >= 90:
		message = "[Excellent] Service health is optimal. All systems are operating within normal parameters."
	case healthScore >= 85:
		message = "[Good] Service health is performing well with minor optimizations recommended."
	case healthScore >= 70:
		message = "[Satisfactory] Service health is stable with room for performance improvements."
	case healthScore >= 50:
		message = "[Fair] Service health is functional but requires attention to resource utilization."
	case healthScore >= 30:
		message = "[Poor] Service health is degraded. Immediate investigation and remediation required."
	default:
		message = "[Critical] Service health is severely compromised. Urgent intervention necessary."
	}

	return message
}

// GetServiceHealth retrieves the service health statistics.
func GetServiceHealth(serviceStats *models.ServiceStats) models.ServiceHealth {
	healthInPercent, err := CalculateHealthScore(serviceStats)
	if err != nil {
		return models.ServiceHealth{
			SystemHealth:  models.Health{Percent: 0, Healthy: false, Message: "Error: Unable to calculate health score. Please check system configuration."},
			ServiceHealth: models.Health{Percent: 0, Healthy: false, Message: "Error: Unable to calculate health score. Please check system configuration."},
		}
	}

	var healthData models.ServiceHealth
	healthData.ServiceHealth.Percent = healthInPercent.ServiceHealth.Percentage
	healthData.SystemHealth.Percent = healthInPercent.SystemHealth.Percentage

	// serviceHealth := healthData.ServiceHealth.Percent
	// systemHealth := healthData.SystemHealth.Percent

	healthData.ServiceHealth = models.Health{
		Percent: healthData.ServiceHealth.Percent,
		Healthy: healthData.ServiceHealth.Percent > 50,
		Message: getStatusMessage(healthData.ServiceHealth.Percent),
		IconMsg: healthInPercent.ServiceHealth.Message,
	}
	healthData.SystemHealth = models.Health{
		Percent: healthData.SystemHealth.Percent,
		Healthy: healthData.SystemHealth.Percent > 50,
		Message: getStatusMessage(healthData.SystemHealth.Percent),
		IconMsg: healthInPercent.SystemHealth.Message,
	}
	return healthData
}

// ConstructRawMemStats constructs a list of raw memory statistics records.
func ConstructRawMemStats(memStats *runtime.MemStats) []models.RawMemStatsRecords {
	r := []models.RawMemStatsRecords{
		newRawRecord("alloc", float64(memStats.Alloc)),
		newRawRecord("total_alloc", float64(memStats.TotalAlloc)),
		newRawRecord("sys", float64(memStats.Sys)),
		newRawRecord("sys", float64(memStats.Sys)),
		newRawRecord("lookups", float64(memStats.Lookups)),
		newRawRecord("mallocs", float64(memStats.Mallocs)),
		newRawRecord("frees", float64(memStats.Frees)),
		newRawRecord("heap_alloc", float64(memStats.HeapAlloc)),
		newRawRecord("heap_sys", float64(memStats.HeapSys)),
		newRawRecord("heap_idle", float64(memStats.HeapIdle)),
		newRawRecord("heap_inuse", float64(memStats.HeapInuse)),
		newRawRecord("heap_released", float64(memStats.HeapReleased)),
		newRawRecord("heap_objects", float64(memStats.HeapObjects)),
		newRawRecord("stack_inuse", float64(memStats.StackInuse)),
		newRawRecord("stack_sys", float64(memStats.StackSys)),
		newRawRecord("m_span_inuse", float64(memStats.MSpanInuse)),
		newRawRecord("m_span_sys", float64(memStats.MSpanSys)),
		newRawRecord("m_cache_inuse", float64(memStats.MCacheInuse)),
		newRawRecord("m_cache_sys", float64(memStats.MCacheSys)),
		newRawRecord("buck_hash_sys", float64(memStats.BuckHashSys)),
		newRawRecord("gc_sys", float64(memStats.GCSys)),
		newRawRecord("other_sys", float64(memStats.OtherSys)),
		newRawRecord("next_gc", float64(memStats.NextGC)),
		newRawRecord("last_gc", float64(memStats.LastGC)),
		newRawRecord("pause_total_ns", float64(memStats.PauseTotalNs)),
		newRawRecord("num_gc", float64(memStats.NumGC)),
		newRawRecord("num_forced_gc", float64(memStats.NumForcedGC)),
		newRawRecord("gc_cpu_fraction", float64(memStats.GCCPUFraction)),
	}

	return r
}

// newRawRecord creates a new Record with appropriate units and human-readable formats.
func newRawRecord(name string, value float64) models.RawMemStatsRecords {
	return models.RawMemStatsRecords{
		RecordName:  name,
		RecordValue: common.ConvertBytesToUnit(value, "KB"),
	}
}
