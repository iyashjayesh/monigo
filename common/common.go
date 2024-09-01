package common

import (
	"fmt"
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

func GetProcessId() int32 {
	return int32(os.Getpid())
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
