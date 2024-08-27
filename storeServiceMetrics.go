package monigo

import (
	"encoding/json"
	"errors"

	bolt "go.etcd.io/bbolt"
)

// ServiceMetrics is the interface for storing and viewing metrics
type ServiceMetricsStore interface {
	StoreServiceRuntimeMetrics(serviceMetrics *ServiceMetrics) error
	GetServiceMetricsFromMonigoDb() (*ServiceMetrics, error)
}

func (db *DBWrapper) StoreServiceRuntimeMetrics(serviceMetrics *ServiceMetrics) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(metricsInfoBucket))
		if err != nil {
			return err
		}

		rowData, err := json.Marshal(serviceMetrics)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(serviceMetrics.Id.String()), rowData)
	})
}

func (db *DBWrapper) GetServiceMetricsFromMonigoDb() ([]ServiceMetrics, error) {
	var serviceMetrics []ServiceMetrics

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(metricsInfoBucket))
		if b == nil {
			return errors.New("bucket not found")
		}

		return b.ForEach(func(k, v []byte) error {
			var sm ServiceMetrics
			if err := json.Unmarshal(v, &sm); err != nil {
				return err
			}
			serviceMetrics = append(serviceMetrics, sm)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return serviceMetrics, nil
}
