package timeseries

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/models"
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

// TimeSeriesServiceMetrics represents metrics for a time series service.
type TimeSeriesServiceMetrics struct {
	Load                   float64
	Cores                  float64
	MemoryUsed             float64
	NumberOfReqServerd     float64
	GoRoutines             float64
	TotalAlloc             float64
	MemoryAllocSys         float64
	HeapAlloc              float64
	HeapAllocSys           float64
	UpTime                 time.Duration
	TotalDurationTookByAPI time.Duration
}

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
			tstorage.WithDataPath("./data"),
			tstorage.WithRetention(1*time.Hour),
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

// StoreServiceMetrics stores the service metrics in the storage.
func StoreServiceMetrics(serviceMetrics *models.TimeSeriesServiceMetrics) error {
	sto, err := GetStorageInstance()
	if err != nil {
		return fmt.Errorf("error getting storage instance: %w", err)
	}

	var rows []tstorage.Row
	timestamp := time.Now().Unix()
	rows = []tstorage.Row{
		{
			Metric:    "load_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.Load},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "cores_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.Cores},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
	}

	if err := sto.InsertRows(rows); err != nil {
		return fmt.Errorf("error storing service metrics: %w", err)
	}

	log.Println("Service metrics stored successfully in the storage, timestamp:", timestamp)
	return nil
}

// GetDataPoints retrieves data points for a given metric and labels.
func GetDataPoints(metric string, labels []tstorage.Label, start, end int64) ([]*tstorage.DataPoint, error) {
	sto, err := GetStorageInstance()
	if err != nil {
		return nil, fmt.Errorf("error getting storage instance: %w", err)
	}
	return sto.Select(metric, labels, start, end)
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

// PurgeStorage removes all storage data and closes the storage.
func PurgeStorage() {
	CloseStorage()
	basePath := GetBasePath()
	log.Println("Purging storage from path:", basePath)
	if err := os.RemoveAll(basePath); err != nil {
		log.Fatalf("Error purging storage: %v\n", err)
	}

	if err := os.RemoveAll("./data"); err != nil {
		log.Fatalf("Error purging storage: %v\n", err)
	}
}
