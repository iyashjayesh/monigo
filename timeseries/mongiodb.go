package timeseries

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/nakabonne/tstorage"
)

// Storage defines the methods required for storage operations.
type Storage interface {
	InsertRows(rows []tstorage.Row) error
	Select(metric string, labels []tstorage.Label, start, end int64) ([]*tstorage.DataPoint, error)
	Close() error
}

var (
	once      sync.Once
	basePath  string
	storage   Storage
	closeOnce sync.Once
)

// StorageWrapper wraps the tstorage.Storage to implement the Storage interface.
type StorageWrapper struct {
	storage tstorage.Storage
}

// InsertRows inserts rows into the storage.
func (s *StorageWrapper) InsertRows(rows []tstorage.Row) error {
	return s.storage.InsertRows(rows)
}

// Select retrieves data points from the storage.
func (s *StorageWrapper) Select(metric string, labels []tstorage.Label, start, end int64) ([]*tstorage.DataPoint, error) {
	return s.storage.Select(metric, labels, start, end)
}

// Close closes the storage connection.
func (s *StorageWrapper) Close() error {
	return s.storage.Close()
}

// GetStorageInstance initializes and returns a Storage instance.
func GetStorageInstance() (Storage, error) {
	var err error
	once.Do(func() {
		basePath = GetBasePath()
		tstorageInstance, err := tstorage.NewStorage(
			tstorage.WithDataPath(basePath+"/data"),
			tstorage.WithRetention(common.GetRetentionPeriod()),
		)
		if err != nil {
			log.Fatalf("Error initializing storage: %v\n", err)
		}
		storage = &StorageWrapper{storage: tstorageInstance}
	})
	return storage, err
}

// GetBasePath returns the base path for storage.
func GetBasePath() string {
	const monigoFolder string = "monigo"

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

// CloseStorage closes the storage instance.
func CloseStorage() {
	closeOnce.Do(func() {
		if storage != nil {
			log.Println("Closing storage instance")
			if err := storage.Close(); err != nil {
				log.Printf("Error closing storage: %v\n", err)
			}
		}
	})
}

// PurgeStorage removes alqqql storage data and closes the storage.
func PurgeStorage() {
	basePath := GetBasePath()
	log.Println("Purging storage from path:", basePath)
	if err := os.RemoveAll(basePath); err != nil {
		log.Fatalf("Error purging storage: %v\n", err)
	}

	if err := os.RemoveAll("./data"); err != nil {
		log.Fatalf("Error purging storage: %v\n", err)
	}
}

func SetDataPointsSyncFrequency(frequency ...string) {
	freqStr := "5m"
	if len(frequency) > 0 {
		freqStr = frequency[0]
	}

	freqTime, err := time.ParseDuration(freqStr)
	if err != nil {
		log.Printf("Invalid frequency format: %v. Using default of 5m.\n", err)
		freqTime = 5 * time.Minute
	}

	freqOnce := sync.Once{}
	freqOnce.Do(func() {
		// serviceMetrics := core.GetServiceMetricsModel()
		serviceMetrics := core.GetServiceStats()
		if err := StoreNewServiceMetrics(&serviceMetrics); err != nil {
			log.Panicf("Error storing service metrics: %v\n", err)
		}
	})

	timer := time.NewTimer(freqTime)
	go func() {
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				// serviceMetrics := core.GetServiceMetricsModel()
				serviceMetrics := core.GetServiceStats()
				if err := StoreNewServiceMetrics(&serviceMetrics); err != nil {
					log.Panicf("Error storing service metrics: %v\n", err)
				}
				size := common.GetDirSize(basePath + "/data")
				log.Println("Size of data directory: ", size)
				timer.Reset(freqTime)
			}
		}
	}()
}
