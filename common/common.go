package common

import (
	"fmt"
	"os"
	"path/filepath"
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

// SetServiceInfo sets the service info.
func SetServiceInfo(serviceName string, serviceStartTime time.Time, goVersion string) {
	serviceInfo.ServiceName = serviceName
	serviceInfo.ServiceStartTime = serviceStartTime
	serviceInfo.GoVersion = goVersion
}

// GetServiceInfo returns the service info.
func GetServiceInfo() models.ServiceInfo {
	return serviceInfo
}
