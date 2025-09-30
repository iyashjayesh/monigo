package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iyashjayesh/monigo/models"
)

const monigoFolder string = "monigo"

var (
	serviceInfo      models.ServiceInfo
	rententionPeriod string
)

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
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			log.Printf("[MoniGo] Warning: failed to create cache directory %s: %v", path, err)
			// Fallback to a temporary directory if we can't create in current directory
			tempDir := os.TempDir()
			path = fmt.Sprintf("%s/%s", tempDir, monigoFolder)
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				log.Printf("[MoniGo] Warning: failed to create fallback cache directory %s: %v", path, err)
			}
		}
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
func SetServiceInfo(serviceName string, serviceStartTime time.Time, goVersion string, processId int32, rentainPeriod string) {
	serviceInfo.ServiceName = serviceName
	serviceInfo.ServiceStartTime = serviceStartTime
	serviceInfo.GoVersion = goVersion
	serviceInfo.ProcessId = processId
	rententionPeriod = rentainPeriod
}

// GetServiceInfo returns the service info.
func GetServiceInfo() models.ServiceInfo {
	return serviceInfo
}

// BytesToGB converts bytes to GB.
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
		"service_memory_load": "Service Memory[RAM] Load is the memory usage of the service",
		"system_memory_load": "System Memory[RAM] Load is the memory usage of the system",
		"total_memory_load": "Total Memory[RAM] Load is the memory usage of the system and the service",
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
		"total_system_memory": "Total System Memory[RAM] is the total memory the system has",
		"memory_used_by_system": "Memory[RAM] Used by System is the memory the system is using",
		"memory_used_by_service": "Memory[RAM] Used by Service is the memory the service is using",
		"available_memory": "Available Memory[RAM] is the memory available on the system",
		"uptime": "Uptime is the time the service has been running",
		"timestamp": "Timestamp is the time the data was collected"
	}`

	var serviceInfo map[string]string
	_ = json.Unmarshal([]byte(data), &serviceInfo)
	return serviceInfo
}

// ParseStringToFloat64 converts string to float64.
func ParseStringToFloat64(value string) float64 {
	var result float64
	fmt.Sscanf(value, "%f", &result)
	return result
}

// RoundFloat64 rounds the float64 value to the specified precision.
func RoundFloat64(value float64, precision int) float64 {
	return ParseStringToFloat64(fmt.Sprintf("%."+fmt.Sprintf("%d", precision)+"f", value))
}

// ConvertToReadableUnit converts the input value to a more human-readable unit.
func ConvertToReadableUnit(value interface{}) string {
	var num float64

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num = float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num = float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		num = v.Float()
	case reflect.String:
		num = ParseStringToFloat64(v.String())
	default:
		log.Panic("[MoniGo] unsupported type: ", v.Kind())
		return ""
	}

	var unit string
	switch {
	case num < 1024:
		unit = "B" // Bytes
	case num < math.Pow(1024, 2): // Less than 1 MB
		unit = "KB" // Kilobytes
		num = num / 1024
	case num < math.Pow(1024, 3): // Less than 1 GB
		unit = "MB" // Megabytes
		num = num / math.Pow(1024, 2)
	case num < math.Pow(1024, 4): // Less than 1 TB
		unit = "GB" // Gigabytes
		num = num / math.Pow(1024, 3)
	case num < math.Pow(1024, 5): // Less than 1 PB
		unit = "TB" // Terabytes
		num = num / math.Pow(1024, 4)
	default:
		unit = "PB" // Petabytes
		num = num / math.Pow(1024, 5)
	}

	return fmt.Sprintf("%.2f %s", num, unit)
}

// BytesToUnit converts a float64 value (representing bytes) to a human-readable unit (KB, MB, GB, TB) based on its magnitude
func BytesToUnit(value uint64) string {
	num := float64(value)
	var unit string
	switch {
	case num < 1024:
		unit = "B" // Bytes
	case num < math.Pow(1024, 2): // Less than 1 MB
		unit = "KB" // Kilobytes
		num = num / 1024
	case num < math.Pow(1024, 3): // Less than 1 GB
		unit = "MB" // Megabytes
		num = num / math.Pow(1024, 2)
	case num < math.Pow(1024, 4): // Less than 1 TB
		unit = "GB" // Gigabytes
		num = num / math.Pow(1024, 3)
	default:
		unit = "TB" // Terabytes
		num = num / math.Pow(1024, 4)
	}

	return fmt.Sprintf("%.2f %s", num, unit)
}

// ConvertBytesToUnit converts bytes to the specified unit and returns the result as a float64.
func ConvertBytesToUnit(bytes float64, unit string) float64 {
	var result float64
	base := float64(1000)
	unit = strings.ToUpper(unit)
	switch unit {
	case "KB":
		result = bytes / base
	case "MB":
		result = bytes / (base * base)
	case "GB":
		result = bytes / (base * base * base)
	case "TB":
		result = bytes / (base * base * base * base)
	default:
		fmt.Println("Unknown unit")
		return 0
	}

	return result
}

// GetServiceStartTime returns the service start time.
func GetServiceStartTime() time.Time {
	return serviceInfo.ServiceStartTime
}

// parseDuration parses the duration string.
func parseDuration(input string) (time.Duration, error) {
	if strings.HasSuffix(input, "d") {
		daysStr := strings.TrimSuffix(input, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, err
		}
		// Convert days to hours
		return time.Duration(days*24) * time.Hour, nil
	} else if strings.HasSuffix(input, "month") {
		monthsStr := strings.TrimSuffix(input, "month")
		months, err := strconv.Atoi(monthsStr)
		if err != nil {
			return 0, err
		}
		return time.Duration(months*30) * 24 * time.Hour, nil
	}

	return time.ParseDuration(input)
}

// GetDataRetentionPeriod returns the retention period.
func GetDataRetentionPeriod() time.Duration {

	if rententionPeriod == "" {
		rententionPeriod = "7d"
	}

	rententionPeriod, err := parseDuration(rententionPeriod)
	if err != nil {
		log.Printf("[MoniGo] Error parsing retention period, using default retention period (7d): %v", err)
		rententionPeriod = time.Duration(7) * 24 * time.Hour
	}

	return rententionPeriod
}

// Helper function to set default string value
func DefaultIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

// Helper function to set default float value
func DefaultFloatIfZero(val, def float64) float64 {
	if val == 0 {
		return def
	}
	return val
}

// Helper function to set default integer value
func DefaultIntIfZero(val, def int) int {
	if val == 0 {
		return def
	}
	return val
}

// Helper function to convert values to MB
func ConvertToMB(value string) (float64, error) {
	value = strings.TrimSpace(value)
	value = strings.Replace(value, " ", "", -1)
	unit := strings.ToUpper(value[len(value)-2:])
	val, err := strconv.ParseFloat(value[:len(value)-2], 64)
	if err != nil {
		return 0, err
	}

	unit = strings.ToUpper(unit)
	switch unit {
	case "TB":
		return val * 1024 * 1024, nil
	case "GB":
		return val * 1024, nil
	case "MB":
		return val, nil
	case "KB":
		return val / 1024, nil
	default:
		return 0, fmt.Errorf("unsupported memory unit: %s", unit)
	}
}

// Cache is the struct to store the cache data
type Cache struct {
	Data map[string]time.Time
}

// SaveToFile saves the cache to a file
func (c *Cache) SaveToFile(filename string) error {
	jsonData, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}

	base64Data := base64.StdEncoding.EncodeToString(jsonData)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(base64Data)
	return err
}

// LoadFromFile loads the cache from a file, or starts fresh if the file does not exist or an error occurs
func (c *Cache) LoadFromFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	if stat, err := file.Stat(); err != nil {
		return fmt.Errorf("failed to stat cache file: %w", err)
	} else if stat.Size() == 0 {
		return nil
	}

	base64Data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	jsonData, err := base64.StdEncoding.DecodeString(string(base64Data))
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	if err := json.Unmarshal(jsonData, &c.Data); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}
