package monigodb

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/iyashjayesh/monigo/models"
	bolt "go.etcd.io/bbolt"
)

type ServiceQueries interface {
	GetServiceName() string
}

func (db *DBWrapper) GetServiceDetails() models.ServiceInfo {
	var serviceData models.ServiceInfo
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(service_info))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		bucket.ForEach(func(k, v []byte) error {
			if err := json.Unmarshal(v, &serviceData); err != nil {
				return err
			}
			return nil
		})

		return nil
	})
	return serviceData
}

func ShowRuntimeMetrics() {
	// Store the Service Metrics
	metricsInfo, err := dbObj.GetServiceMetricsFromMonigoDb()
	if err != nil {
		log.Fatalf("Error getting service metrics: %v\n", err)
	}

	log.Printf("Service Metrics: %+v\n", metricsInfo)
}
