package monigo

import (
	"log"

	bolt "go.etcd.io/bbolt"
)

const (
	// bucket name
	serviceInfoBucket        = "service_info"
	metricsInfoBucket        = "metrics_info"
	runtimeMetricsInfoBucket = "runtime_metrics_info"
)

// connectDb opens the database and returns a pointer to the bolt.DB instance
func connectDb() (*DBWrapper, error) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	basePath := getBasePath()
	db, err := bolt.Open(basePath+"/monigo.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	return &DBWrapper{db}, nil
}

// closeDb closes the bolt.DB instance
func closeDb(db *bolt.DB) {
	if err := db.Close(); err != nil {
		log.Fatal("Failed to close database:", err)
	}
}
