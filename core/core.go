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

func GetServiceStats() models.ServiceStats {

	var serviceStats models.ServiceStats
	// timeNow := time.Now()
	serviceStats.CoreStatistics = GetCoreStatistics()
	// log.Println("Time taken to get the core statistics: ", time.Since(timeNow))

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		// timeNow := time.Now()
		serviceStats.LoadStatistics = GetLoadStatistics()
		// log.Println("Time taken to get the load statistics: ", time.Since(timeNow))
	}()

	go func() {
		defer wg.Done()
		// timeNow := time.Now()
		serviceStats.MemoryStatistics = GetMemoryStatistics()
		// log.Println("Time taken to get the memory statistics: ", time.Since(timeNow))
	}()

	go func() {
		defer wg.Done()
		// timeNow := time.Now()
		serviceStats.CPUStatistics = GetCPUStatistics()
		// log.Println("Time taken to get the CPU statistics: ", time.Since(timeNow))
	}()

	go func() {
		defer wg.Done()
		// timeNow := time.Now()
		memStats := ReadMemStats()
		serviceStats.HeapAllocByService = common.BytesToUnit(memStats.HeapAlloc)
		serviceStats.HeapAllocBySystem = common.BytesToUnit(memStats.HeapSys)
		serviceStats.TotalAllocByService = common.BytesToUnit(memStats.TotalAlloc)
		serviceStats.TotalMemoryByOS = common.BytesToUnit(memStats.Sys)
		// log.Println("Time taken to get the memory stats: ", time.Since(timeNow))
	}()

	go func() {
		defer wg.Done()
		// timeNow := time.Now()
		serviceStats.NetworkIO.BytesReceived, serviceStats.NetworkIO.BytesSent = GetNetworkIO()
		// log.Println("Time taken to get the network stats: ", time.Since(timeNow))
	}()

	wg.Wait()

	serviceStats.Health = GetServiceHealth(&serviceStats.LoadStatistics)
	// serviceStats.DiskIO = GetDiskIO()                                             // TODO: Need to implement this function

	return serviceStats
}

func GetCoreStatistics() models.CoreStatistics {
	rcount, durationTook := GetServiceMetrics()

	serviceInfo := common.GetServiceInfo()

	uptime := time.Since(serviceInfo.ServiceStartTime)
	uptimeStr := fmt.Sprintf("%.2f s", uptime.Seconds())

	// Formatting uptime based on its duration
	if uptime.Seconds() > 60 {
		uptimeStr = fmt.Sprintf("%.2f m", uptime.Minutes())
	}
	if uptime.Minutes() > 60 {
		uptimeStr = fmt.Sprintf("%.2f h", uptime.Hours())
	}
	if uptime.Hours() > 24 {
		uptimeStr = fmt.Sprintf("%.2f d", uptime.Hours()/24)
	}
	if uptime.Hours() > 30*24 {
		uptimeStr = fmt.Sprintf("%.2f mo", uptime.Hours()/(30*24))
	}
	if uptime.Hours() > 12*30*24 {
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
		RawMemStatsRecords:  ConstructRawMemStats(m),
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

// GetServiceHealth retrieves the service health statistics.
func GetServiceHealth(loadStats *models.LoadStatistics) models.ServiceHealth {
	threshold := GetServiceHealthThresholdsModel()

	healthInPercent, err := CalculateHealthScore(threshold)
	if err != nil {
		log.Println("Error calculating health score:", err)
		return models.ServiceHealth{
			SystemHealth:  models.Health{Percent: 0, Healthy: false, Message: "Oops! We hit a snag while calculating the health score."},
			ServiceHealth: models.Health{Percent: 0, Healthy: false, Message: "Oops! We hit a snag while calculating the health score."},
			OverallHealth: models.Health{Percent: 0, Healthy: false, Message: "Oops! We hit a snag while calculating the health score."},
		}
	}

	var healthData models.ServiceHealth
	healthData.ServiceHealth.Percent = healthInPercent.ServiceHealth
	healthData.SystemHealth.Percent = healthInPercent.SystemHealth
	healthData.OverallHealth.Percent = healthInPercent.OverallHealth

	overallHealth := healthInPercent.OverallHealth
	healthy := overallHealth > 50
	var message string

	switch {
	case overallHealth >= 90:
		message = "[Outstanding] The Overall Health is rocking it! Everything’s running smoothly and life is good."
	case overallHealth >= 85:
		message = "[Impressive] The Overall Health is doing great—just a few hiccups that need a tweak here and there."
	case overallHealth >= 70:
		message = "[Solid] The Overall Health is holding up well. A bit of fine-tuning could make it shine even brighter."
	case overallHealth >= 50:
		message = "[Fair] The Overall Health is functional but could use a bit of TLC. Time to check those resources!"
	case overallHealth >= 30:
		message = "[Wobbly] The Overall Health is feeling the heat. Roll up your sleeves and dig into those logs!"
	default:
		message = "[Oops] The Overall Health is in rough shape. Time to call in the cavalry and get things back on track!"
	}

	healthData.ServiceHealth.Healthy = healthy
	healthData.ServiceHealth.Message = message
	healthData.SystemHealth.Healthy = healthy
	healthData.SystemHealth.Message = message
	healthData.OverallHealth.Healthy = healthy
	healthData.OverallHealth.Message = message

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
