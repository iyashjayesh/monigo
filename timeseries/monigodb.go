package timeseries

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/nakabonne/tstorage"
)

var (
	once      sync.Once          // Ensures that the storage is initialized only once
	basePath  string             // Base path for storage
	storage   Storage            // Storage instance
	closeOnce sync.Once          // Ensures that the storage is closed only once
	ctx       context.Context    // Context for goroutines
	cancel    context.CancelFunc // Cancel function for goroutines
)

// Storage defines the methods required for storage operations.
type Storage interface {
	InsertRows(rows []tstorage.Row) error
	Select(metric string, labels []tstorage.Label, start, end int64) ([]*tstorage.DataPoint, error)
	Close() error
}

// StorageWrapper wraps the tstorage.Storage to implement the Storage interface.
type StorageWrapper struct {
	storage tstorage.Storage
	closed  bool
	mu      sync.Mutex
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil // or return a specific error indicating it's already closed
	}

	s.closed = true
	return s.storage.Close()
}

// GetStorageInstance initializes and returns a Storage instance.
func GetStorageInstance() (Storage, error) {
	var err error
	once.Do(func() {
		basePath = common.GetBasePath()
		storageInstance, initErr := tstorage.NewStorage(
			tstorage.WithDataPath(basePath+"/data"),
			tstorage.WithRetention(common.GetDataRetentionPeriod()),
		)
		if initErr != nil {
			err = initErr
			log.Printf("[MoniGo] Error initializing storage: %v\n", err)
			return
		}
		storage = &StorageWrapper{storage: storageInstance}
		// Initialize context and cancel function for goroutines
		ctx, cancel = context.WithCancel(context.Background())
	})
	return storage, err
}

// CloseStorage closes the storage instance and stops any running goroutines.
func CloseStorage() error {
	var err error
	closeOnce.Do(func() {
		if cancel != nil {
			cancel() // Stop any goroutines
		}
		if storage != nil {
			if closeErr := storage.Close(); closeErr != nil {
				log.Printf("[MoniGo] Error closing storage: %v\n", closeErr)
				err = closeErr
			}
		}
	})
	return err
}

// PurgeStorage removes all storage data and closes the storage.
func PurgeStorage() error {
	basePath := common.GetBasePath()
	if err := os.RemoveAll(basePath); err != nil {
		log.Printf("[MoniGo] Error purging storage: %v\n", err)
		return err
	}
	return nil
}

// SetDataPointsSyncFrequency sets the frequency at which data points are synchronized.
func SetDataPointsSyncFrequency(frequency ...string) error {
	freqStr := "5m"
	if len(frequency) > 0 {
		freqStr = frequency[0]
	}

	freqTime, err := time.ParseDuration(freqStr)
	if err != nil {
		log.Printf("[MoniGo] Invalid frequency format: %v. Using default of 5m.\n", err)
		freqTime = 5 * time.Minute
	}

	// Initializing service metrics once
	serviceMetrics := core.GetServiceStats()
	if err := StoreServiceMetrics(&serviceMetrics); err != nil {
		return errors.New("[MoniGo] error storing service metrics, err: " + err.Error())
	}

	timer := time.NewTimer(freqTime)
	go func() {
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				serviceMetrics := core.GetServiceStats()
				if err := StoreServiceMetrics(&serviceMetrics); err != nil {
					log.Printf("[MoniGo] Error storing service metrics: %v\n", err)
				}
				timer.Reset(freqTime)
			}
		}
	}()

	return nil
}
