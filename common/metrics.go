package common

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

// GetCPULoad calculates the CPU load for the service, system, and total.
func GetCPULoad() (serviceCPU, systemCPU, totalCPU string) {

	proc := GetProcessObject()            // Getting process details
	serviceCPUF, err := proc.CPUPercent() // 	Measure CPU percent for the current process
	if err != nil {
		log.Panicf("[MoniGo] Error fetching CPU load for the service: %v\n", err)
	}
	serviceCPU = ParseFloat64ToString(serviceCPUF) + "%" // Service CPU usage percentage

	cpuPercents, err := cpu.Percent(time.Second, false) // Get total system CPU percentage
	if err != nil {
		log.Panicf("[MoniGo] Error fetching CPU load for the system: %v\n", err)
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
		log.Panicf("[MoniGo] Error fetching memory load for the system: %v\n", err)
	}
	systemMem = ParseFloat64ToString(vmStat.UsedPercent) + "%"          // Calculate system memory as a percentage of total memory
	totalMem = ParseFloat64ToString(ParseUint64ToFloat64(vmStat.Total)) // Total memory in bytes Total amount of RAM on this system

	proc := GetProcessObject()
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		log.Panicf("[MoniGo] Error fetching memory load for the service: %v\n", err)
	}

	serviceMem = ParseFloat64ToString(float64(memInfo.RSS)/float64(vmStat.Total)*100) + "%" // Calculate service memory as a percentage of total memory

	return serviceMem, systemMem, totalMem
}

// GetProcessDetails returns the process ID and process object.
func GetProcessDetails() (int32, *process.Process) {
	pid := GetProcessId()
	proc, err := process.NewProcess(pid)
	if err != nil {
		log.Panicf("[MoniGo] Error fetching process details: %v\n", err)
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
