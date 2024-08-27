package monigo

import (
	"encoding/json"
	"errors"

	bolt "go.etcd.io/bbolt"
)

// DBWrapper is a wrapper around bbolt.DB that allows us to define methods
type DBWrapper struct {
	*bolt.DB
}

// MetricsStore is the interface for storing and viewing metrics
type MetricsStore interface {
	StoreServiceInfo(storeServiceInfo *ServiceInfo) error
	GetServiceInfo(serviceName string) (*ServiceInfo, error)
}

// StoreServiceInfo stores the service metrics in BoltDB
func (db *DBWrapper) StoreServiceInfo(storeServiceInfo *ServiceInfo) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Create or get the bucket for the service info
		bucket, err := tx.CreateBucketIfNotExists([]byte(serviceInfoBucket))
		if err != nil {
			return err
		}

		// Serialize the row data to JSON
		rowData, err := json.Marshal(storeServiceInfo)
		if err != nil {
			return err
		}

		// Store the row data in the bucket with the service name as the key
		return bucket.Put([]byte(storeServiceInfo.ServiceName), rowData)
	})
}

// GetServiceInfo retrieves the service info from BoltDB
func (db *DBWrapper) GetServiceInfo(serviceName string) (*ServiceInfo, error) {
	var serviceInfo ServiceInfo

	err := db.View(func(tx *bolt.Tx) error {
		// Get the bucket for the service info
		bucket := tx.Bucket([]byte(serviceInfoBucket))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		// Get the row data from the bucket using the service name as the key
		rowData := bucket.Get([]byte(serviceName))
		if rowData == nil {
			return errors.New("service info not found")
		}

		// Deserialize the row data into ServiceInfo
		return json.Unmarshal(rowData, &serviceInfo)
	})

	return &serviceInfo, err
}
