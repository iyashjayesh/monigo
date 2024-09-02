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
func GetCPULoad() (serviceCPU, systemCPU, totalCPU float64) {

	cpuPercents, err := cpu.Percent(time.Second, false) // Get total system CPU percentage
	if err != nil {
		log.Panicf("Error fetching CPU load for the system: %v\n", err)
	}
	if len(cpuPercents) > 0 {
		systemCPU = cpuPercents[0] // This is the overall system CPU usage percentage.
	}

	proc := GetProcessObject() // Getting process details

	serviceCPU, err = proc.CPUPercent() // 	Measure CPU percent for the current process
	if err != nil {
		log.Panicf("Error fetching CPU load for the service: %v\n", err)
	}

	totalCPU = systemCPU

	return serviceCPU, systemCPU, totalCPU
}

// GetMemoryLoad calculates the memory load for the service, system, and total.
func GetMemoryLoad() (serviceMem, systemMem, totalMem float64) {
	// Get system memory statistics
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Panicf("Error fetching memory load for the system: %v\n", err)
	}
	systemMem = vmStat.UsedPercent

	totalMem = ParseUint64ToFloat64(vmStat.Total)
	proc := GetProcessObject()
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		log.Panicf("Error fetching memory load for the service: %v\n", err)
	}
	serviceMem = float64(memInfo.RSS) / float64(vmStat.Total) * 100 // Calculate service memory as a percentage of total memory

	return serviceMem, systemMem, totalMem
}

// func GetDiskLoad() (serviceDisk, systemDisk, totalDisk float64) {
// 	// Get system disk usage statistics
// 	diskStat, err := disk.Usage("/")
// 	if err != nil {
// 		log.Panicf("Error fetching disk load for the system: %v\n", err)
// 	}
// 	systemDisk = diskStat.UsedPercent
// 	totalDisk = float64(diskStat.Total)
// 	proc := GetProcessObject()

// 	ioStat, err := proc.IOCounters()
// 	if err != nil {
// 		log.Panicf("Error fetching disk load for the service: %v\n", err)
// 	}

// 	// Service disk usage can be considered as the total bytes read and written by the service process
// 	serviceDisk = float64(ioStat.ReadBytes + ioStat.WriteBytes)

// 	return serviceDisk, systemDisk, totalDisk
// }

func GetProcessDetails() (int32, *process.Process) {
	pid := GetProcessId()
	proc, err := process.NewProcess(pid)
	if err != nil {
		log.Panicf("Error fetching process details: %v\n", err)
	}
	return pid, proc
}

func GetProcessId() int32 {
	return int32(os.Getpid())
}

func GetProcessObject() *process.Process {
	_, proc := GetProcessDetails()
	return proc
}

func ParseUint64ToFloat64(value uint64) float64 {
	return float64(value)
}

func ParseFloat64ToString(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func GetVirtualMemory() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}
