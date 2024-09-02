package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iyashjayesh/monigo/models"
)

const monigoFolder string = "monigo"

var serviceInfo models.ServiceInfo

// GetBasePath returns the base path for storage.
func GetBasePath() string {
	var path string
	appPath, _ := os.Getwd()
	if appPath == "/" {
		path = fmt.Sprintf("%s%s", appPath, monigoFolder)
	} else {
		path = fmt.Sprintf("%s/%s", appPath, monigoFolder)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	return path
}

// GetDirSize returns the size of the directory.
func GetDirSize(folderPath string) string {
	var size int64
	filepath.Walk(folderPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
}

// sets the service info.
func SetServiceInfo(serviceName string, serviceStartTime time.Time, goVersion string, processId int32) {
	serviceInfo.ServiceName = serviceName
	serviceInfo.ServiceStartTime = serviceStartTime
	serviceInfo.GoVersion = goVersion
	serviceInfo.ProcessId = processId
}

// GetServiceInfo returns the service info.
func GetServiceInfo() models.ServiceInfo {
	return serviceInfo
}

func BytesToGB(bytes uint64) float64 {
	return float64(bytes) / 1024 / 1024 / 1024
}

// convertToReadableSize converts bytes into a more human-readable format (MB, GB).
func ConvertToReadableSize(bytes uint64) (float64, string) {
	const (
		KB = 1024
		MB = KB * KB
		GB = MB * KB
	)

	switch {
	case bytes >= GB:
		return float64(bytes) / GB, "GB"
	case bytes >= MB:
		return float64(bytes) / MB, "MB"
	case bytes >= KB:
		return float64(bytes) / KB, "KB"
	default:
		return float64(bytes), "bytes"
	}
}

// ConvertBytes converts bytes to the specified unit.
func ConvertBytes(bytes uint64, unit string) float64 {
	unit = strings.ToUpper(unit) // Handle case-insensitivity
	switch unit {
	case "KB":
		return float64(bytes) / 1024.0
	case "MB":
		return float64(bytes) / 1048576.0
	case "GB":
		return float64(bytes) / 1073741824.0
	case "TB":
		return float64(bytes) / 1099511627776.0
	default: // "B" or unspecified
		return float64(bytes)
	}
}

// ConstructJsonFieldDescription reads the data from the JSON file.
func ConstructJsonFieldDescription() map[string]string {

	data := `{
		"service_name": "Service Name is configured by the user while starting the monigo service",
		"service_port": "Service Port is configured by the user while starting the monigo service",
		"go_version": "Go version is the version the service is running on",
		"process_id": "Process ID is the process id of the monigo service",
		"goroutines": "Goroutines is the number of goroutines running in the service",
		"requests_count": "Requests Count is the number of requests served by the service",
		"service_cpu_load": "Service CPU Load is the CPU usage of the service",
		"system_cpu_load": "System CPU Load is the CPU usage of the system",
		"total_cpu_load": "Total CPU Load is the CPU usage of the system and the service",
		"service_memory_load": "Service Memory Load is the memory usage of the service",
		"system_memory_load": "System Memory Load is the memory usage of the system",
		"total_memory_load": "Total Memory Load is the memory usage of the system and the service",
		"service_disk_load": "Service Disk Load is the disk usage of the service",
		"system_disk_load": "System Disk Load is the disk usage of the system",
		"total_disk_load": "Total Disk Load is the disk usage of the system and the service",
		"overall_load_of_service": "Overall Load of Service is the overall load service is under on the system",
		"total_cores": "Total Cores is the number of cores the system has",
		"total_logical_cores": "Total Logical Cores is the number of logical cores the system has",
		"cores_used_by_system": "Cores Used by System is the number of cores the system is using",
		"cores_used_by_service": "Cores Used by Service is the number of cores the service is using",
		"cores_used_by_service_in_percentage": "Cores Used by Service in Percentage is the percentage of cores the service is using",
		"cores_used_by_system_in_percentage": "Cores Used by System in Percentage is the percentage of cores the system is using",
		"total_system_memory": "Total System Memory is the total memory the system has",
		"memory_used_by_system": "Memory Used by System is the memory the system is using",
		"memory_used_by_service": "Memory Used by Service is the memory the service is using",
		"available_memory": "Available Memory is the memory available on the system",
		"uptime": "Uptime is the time the service has been running",
		"timestamp": "Timestamp is the time the data was collected"
	}`

	var serviceInfo map[string]string
	err := json.Unmarshal([]byte(data), &serviceInfo)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	return serviceInfo
}

func ParseStringToFloat64(value string) float64 {
	var result float64
	fmt.Sscanf(value, "%f", &result)
	return result
}
