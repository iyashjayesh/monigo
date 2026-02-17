package timeseries

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/nakabonne/tstorage"
)

// Storage defines the methods required for storage operations.
type Storage interface {
	InsertRows(rows []Row) error
	Select(metric string, labels []Label, start, end int64) ([]DataPoint, error)
	Close() error
}

// InMemoryStorage provides an in-memory implementation of the Storage interface.
type InMemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]DataPoint
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string][]DataPoint),
	}
}

func (s *InMemoryStorage) InsertRows(rows []Row) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, row := range rows {
		s.data[row.Metric] = append(s.data[row.Metric], row.DataPoint)
	}
	return nil
}

func (s *InMemoryStorage) Select(metric string, labels []Label, start, end int64) ([]DataPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	points, ok := s.data[metric]
	if !ok {
		return nil, nil
	}

	var result []DataPoint
	for _, p := range points {
		if p.Timestamp >= start && p.Timestamp <= end {
			result = append(result, p)
		}
	}
	return result, nil
}

func (s *InMemoryStorage) Close() error {
	return nil
}

// StorageWrapper wraps the tstorage.Storage to implement the Storage interface.
type StorageWrapper struct {
	storage tstorage.Storage
	closed  bool
	mu      sync.Mutex
}

// InsertRows inserts rows into the storage, converting monigo types to tstorage types.
func (s *StorageWrapper) InsertRows(rows []Row) error {
	return s.storage.InsertRows(toTStorageRows(rows))
}

// Select retrieves data points from the storage, converting tstorage types to monigo types.
func (s *StorageWrapper) Select(metric string, labels []Label, start, end int64) ([]DataPoint, error) {
	points, err := s.storage.Select(metric, toTStorageLabels(labels), start, end)
	if err != nil {
		return nil, err
	}
	return fromTStorageDataPoints(points), nil
}

// Close closes the storage connection.
func (s *StorageWrapper) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	return s.storage.Close()
}

type storageManager struct {
	storage   Storage
	ctx       context.Context
	cancel    context.CancelFunc
	once      sync.Once
	closeOnce sync.Once
	mu        sync.Mutex
}

var (
	manager     = &storageManager{}
	storageType = "disk" // "disk" or "memory"
)

// SetStorageType sets the storage type
func SetStorageType(t string) {
	storageType = t
}

// GetStorageInstance initializes and returns a Storage instance.
func GetStorageInstance() (Storage, error) {
	var err error
	manager.once.Do(func() {
		if storageType == "memory" {
			manager.storage = NewInMemoryStorage()
			manager.ctx, manager.cancel = context.WithCancel(context.Background())
			return
		}

		basePath := common.GetBasePath()
		storageInstance, initErr := tstorage.NewStorage(
			tstorage.WithDataPath(filepath.Join(basePath, "data")),
			tstorage.WithRetention(common.GetDataRetentionPeriod()),
		)
		if initErr != nil {
			err = initErr
			logger.Log.Error("initializing storage", "error", err)
			return
		}
		manager.storage = &StorageWrapper{storage: storageInstance}
		// Initialize context and cancel function for goroutines
		manager.ctx, manager.cancel = context.WithCancel(context.Background())
	})
	return manager.storage, err
}

// CloseStorage closes the storage instance and stops any running goroutines.
func CloseStorage() error {
	var err error
	manager.closeOnce.Do(func() {
		if manager.cancel != nil {
			manager.cancel() // Stop any goroutines
		}
		if manager.storage != nil {
			if closeErr := manager.storage.Close(); closeErr != nil {
				logger.Log.Error("closing storage", "error", closeErr)
				err = closeErr
			}
		}
	})
	return err
}

// PurgeStorage removes only the monigo data directory to avoid accidental deletions of other files.
func PurgeStorage() error {
	basePath := common.GetBasePath()

	// Safety check: ensure we are only deleting the 'monigo' directory
	if !strings.HasSuffix(basePath, "monigo") {
		return fmt.Errorf("[MoniGo] Refusing to purge storage: basePath %q does not end with 'monigo'", basePath)
	}

	if err := os.RemoveAll(basePath); err != nil {
		logger.Log.Error("purging storage", "error", err)
		return err
	}

	// Recreate the directory
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		logger.Log.Error("recreating storage directory", "error", err)
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
		logger.Log.Warn("invalid frequency format, using default 5m", "error", err)
		freqTime = 5 * time.Minute
	}

	// Ensure storage is initialized before starting the sync loop
	if _, err := GetStorageInstance(); err != nil {
		return err
	}

	// Initializing service metrics once
	serviceMetrics := core.GetServiceStats(context.Background())
	if err := StoreServiceMetrics(&serviceMetrics); err != nil {
		return errors.New("[MoniGo] error storing service metrics, err: " + err.Error())
	}

	ticker := time.NewTicker(freqTime)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-manager.ctx.Done():
				return
			case <-ticker.C:
				serviceMetrics := core.GetServiceStats(manager.ctx)
				if err := StoreServiceMetrics(&serviceMetrics); err != nil {
					logger.Log.Error("storing service metrics", "error", err)
				}
			}
		}
	}()

	return nil
}
