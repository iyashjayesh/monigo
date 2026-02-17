package core

import (
	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/shirou/gopsutil/disk"
)

// GetDiskIO retrieves the disk I/O statistics (Read/Write bytes).
func GetDiskIO() (uint64, uint64) {
	// fetching IO counters for all disks
	ioCounters, err := disk.IOCounters()
	if err != nil {
		logger.Log.Warn("Error fetching disk I/O statistics", "error", err)
		return 0, 0
	}

	var totalReadBytes, totalWriteBytes uint64
	for _, io := range ioCounters {
		totalReadBytes += io.ReadBytes
		totalWriteBytes += io.WriteBytes
	}

	return totalReadBytes, totalWriteBytes
}
