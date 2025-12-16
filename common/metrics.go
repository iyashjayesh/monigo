package common

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

// GetCPULoad calculates the CPU load for the service, system, and total.
func GetCPULoad() (serviceCPU, systemCPU, totalCPU string) {

	proc := GetProcessObject()            // Getting process details
	serviceCPUF, err := proc.CPUPercent() // 	Measure CPU percent for the current process
	if err != nil {
		log.Printf("[MoniGo] Error fetching CPU load for the service: %v\n", err)
		serviceCPUF = 0
	}
	serviceCPU = ParseFloat64ToString(serviceCPUF) + "%" // Service CPU usage percentage

	cpuPercents, err := cpu.Percent(time.Second, false) // Get total system CPU percentage
	if err != nil {
		log.Printf("[MoniGo] Error fetching CPU load for the system: %v\n", err)
		return serviceCPU, "0%", "0%"
	}
	if len(cpuPercents) > 0 {
		systemCPU = ParseFloat64ToString(cpuPercents[0]-serviceCPUF) + "%" // System CPU usage percentage
	}

	totalCPU = ParseFloat64ToString(serviceCPUF+cpuPercents[0]) + "%" // Total CPU usage percentage
	return serviceCPU, systemCPU, totalCPU
}

// GetMemoryLoad calculates the memory load for the service, system, and total.
func GetMemoryLoad() (serviceMem, systemMem, totalMem string) {
	// Get system memory statistics
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("[MoniGo] Error fetching memory load for the system: %v\n", err)
		return "0%", "0%", "0%"
	}
	systemMem = ParseFloat64ToString(vmStat.UsedPercent) + "%"          // Calculate system memory as a percentage of total memory
	totalMem = ParseFloat64ToString(ParseUint64ToFloat64(vmStat.Total)) // Total memory in bytes Total amount of RAM on this system

	proc := GetProcessObject()
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		log.Printf("[MoniGo] Error fetching memory load for the service: %v\n", err)
		return "0%", systemMem, totalMem
	}

	serviceMem = ParseFloat64ToString(float64(memInfo.RSS)/float64(vmStat.Total)*100) + "%" // Calculate service memory as a percentage of total memory

	return serviceMem, systemMem, totalMem
}

// GetDiskLoad calculates the disk load for the service, system, and total.
func GetDiskLoad() (serviceDisk, systemDisk, totalDisk string) {
	// For disk, "Service" usage handles read/write bytes or handle count, but normally "Load" implies storage usage.
	// However, gathering "Disk Usage by Process" is complex and often requires root or specific tracking.
	// For now, we will track System Disk Usage (Root Partition).

	diskUsage, err := disk.Usage("/")
	if err != nil {
		log.Printf("[MoniGo] Error fetching disk usage: %v\n", err) // Changed from Panic to Printf as agreed in plan
		return "0%", "0%", "0%"
	}

	// ServiceDiskLoad is complex to calculate per process without cgroups/root.
	// We will mistakenly leave it as 0% or maybe revisit if we can get FD count as proxy?
	// For now, let's just return System Disk Usage.

	systemDisk = ParseFloat64ToString(diskUsage.UsedPercent) + "%"
	totalDisk = ParseFloat64ToString(float64(diskUsage.Total)) // Total disk size in bytes

	// ServiceDiskLoad: Not easily available.
	serviceDisk = "0%"

	return serviceDisk, systemDisk, totalDisk
}

// GetProcessDetails returns the process ID and process object.
func GetProcessDetails() (int32, *process.Process) {
	pid := GetProcessId()
	proc, err := process.NewProcess(pid)
	if err != nil {
		log.Printf("[MoniGo] Error fetching process details: %v\n", err)
		return pid, nil
	}
	return pid, proc
}

// GetProcessId returns the process ID.
func GetProcessId() int32 {
	return int32(os.Getpid())
}

// GetProcessObject returns the process object.
func GetProcessObject() *process.Process {
	_, proc := GetProcessDetails()
	return proc
}

// ParseUint64ToFloat64 converts uint64 to float64.
func ParseUint64ToFloat64(value uint64) float64 {
	return float64(value)
}

// ParseFloat64ToString converts float64 to string.
func ParseFloat64ToString(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

// GetVirtualMemory returns the virtual memory statistics.
func GetVirtualMemory() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}
