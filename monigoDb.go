package monigo

import (
	"log"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	// bucket name
	serviceInfoBucket = "service_info"
	metricsInfoBucket = "metrics_info"
	// runtimeMetricsInfoBucket = "runtime_metrics_info"
)

var (
	interval = time.Duration(5) * time.Minute
)

// connectDb opens the database and returns a pointer to the bolt.DB instance
func connectDb() (*DBWrapper, error) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	basePath := GetBasePath()
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

// PurgeMonigoDb removes the monigo.db file
func PurgeMonigoDb() {
	basePath := GetBasePath()
	basePath = basePath + "/monigo/monigo.db"
	if _, err := os.Stat(basePath); err == nil {
		err = os.Remove(basePath)
		if err != nil {
			log.Println("Error removing the monigo.db file: ", err)
		} else {
			log.Println("monigo.db file removed successfully")
		}
	}
}

// SetupMonigoSyncInterval sets the interval for storing the metrics
func SetDbSyncFrequency(intervalStr ...string) {

	if len(intervalStr) > 0 {
		intervalStr[0] = "5m"
	}

	interval, err := time.ParseDuration(intervalStr[0])
	if err != nil {
		log.Panicln("Error parsing the interval: ", err)
	}

	Once.Do(func() {
		log.Println("Syncing Service Info to DB once on startup")
		SyncMetricsInfoToDB()
	})

	timer := time.NewTimer(interval * time.Minute)
	go func() {
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				SyncMetricsInfoToDB()
				log.Println("Next sync in ", interval, " minutes", time.Now().Add(interval*time.Minute).Format("2006-01-02 15:04:05"))
				// Reset the timer and start it again
				if !timer.Stop() {
					<-timer.C // Drain the channel if necessary
				}
				timer.Reset(interval * time.Minute)
			}
		}
	}()
}

func SyncMetricsInfoToDB() {
	// Store the Service Metrics
	log.Println("Syncing Service Metrics to DB")
	dbObj.StoreServiceRuntimeMetrics(GetServiceMetricsFromMonigoDbData())
}
