// Function for recording and retrieving metrics.
package monigo

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var (
	mu              sync.Mutex
	requestCount    int
	totalDuration   time.Duration
	functionMetrics = make(map[string]*FunctionMetrics)
)

func RecordRequestDuration(duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	requestCount++
	totalDuration += duration
}

func GetServiceMetrics() (int, time.Duration, *runtime.MemStats) {
	mu.Lock()
	defer mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return requestCount, totalDuration, &memStats
}

func GetFunctionMetrics(functionName string) *FunctionMetrics {
	mu.Lock()
	defer mu.Unlock()
	return functionMetrics[functionName]
}

func GetProcessSats() ProcessStats {

	pid, proc, err := GetProcessDetails()
	if err != nil {
		fmt.Printf("Error fetching process information: %v\n", err)
		return ProcessStats{}
	}

	// Getting system and process resource usage
	sysCPUPercent, sysMemUsage, err := getSystemUsage()
	if err != nil {
		fmt.Printf("Error fetching system usage: %v\n", err)
		return ProcessStats{}
	}

	procCPUPercent, procMemPercent, err := getProcessUsage(proc, &sysMemUsage)
	if err != nil {
		fmt.Printf("Error fetching process usage: %v\n", err)
		return ProcessStats{}
	}

	totalCores, _ := cpu.Counts(false)
	totalLogicalCores := runtime.NumCPU()
	systemUsedCores := (sysCPUPercent / 100) * float64(totalLogicalCores)
	processUsedCores := (procCPUPercent / 100) * float64(totalLogicalCores)

	return ProcessStats{
		ProcessId:         pid,
		SysCPUPercent:     sysCPUPercent,
		ProcCPUPercent:    procCPUPercent,
		ProcMemPercent:    procMemPercent,
		TotalMemoryUsage:  sysMemUsage.TotalMemoryUsage,
		FreeMemory:        sysMemUsage.FreeMemory,
		UsedMemoryPercent: sysMemUsage.UsedPercent,
		TotalCores:        totalCores,
		TotalLogicalCores: totalLogicalCores,
		SystemUsedCores:   systemUsedCores,
		ProcessUsedCores:  processUsedCores,
	}
}

func GetProcessDetails() (int32, *process.Process, error) {
	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		return 0, nil, err
	}
	return pid, proc, nil
}

// Fetches and returns system CPU and memory usage
func getSystemUsage() (float64, Memory, error) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, Memory{}, err
	}

	memUsage := Memory{}
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0, Memory{}, err
	}

	memUsage = Memory{
		TotalMemoryUsage: float64(memInfo.Total),
		FreeMemory:       float64(memInfo.Free),
		UsedPercent:      memInfo.UsedPercent,
	}

	return cpuPercent[0], memUsage, nil
}

// Fetches and returns process CPU and memory usage
func getProcessUsage(proc *process.Process, sysMemUsage *Memory) (float64, float64, error) {
	procCPUPercent, err := proc.CPUPercent()
	if err != nil {
		return 0, 0, err
	}

	procMem, err := proc.MemoryInfo()
	if err != nil {
		return 0, 0, err
	}

	if sysMemUsage.TotalMemoryUsage == 0 {
		return 0, 0, fmt.Errorf("error fetching system memory usage")
	}

	procMemPercent := float64(procMem.RSS) / sysMemUsage.TotalMemoryUsage * 100

	return procCPUPercent, procMemPercent, nil
}
