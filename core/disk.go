package core

import (
	"log"

	"github.com/shirou/gopsutil/disk"
)

// GetDiskIO retrieves the disk I/O statistics (Read/Write bytes).
func GetDiskIO() (uint64, uint64) {
	// fetching IO counters for all disks
	ioCounters, err := disk.IOCounters()
	if err != nil {
		log.Printf("[MoniGo] Warning: Error fetching disk I/O statistics: %v", err)
		return 0, 0
	}

	var totalReadBytes, totalWriteBytes uint64
	for _, io := range ioCounters {
		totalReadBytes += io.ReadBytes
		totalWriteBytes += io.WriteBytes
	}

	return totalReadBytes, totalWriteBytes
}
