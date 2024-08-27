package monigodb

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/iyashjayesh/monigo/core"
	"github.com/iyashjayesh/monigo/models"
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
	basePath string
	Once     sync.Once = sync.Once{}
	dbObj    *DBWrapper
)

func GetDbInstance() *DBWrapper {
	Once.Do(func() {
		log.Println("Initializing the DB")
		var err error
		dbObj, err = ConnectDb()
		if err != nil {
			log.Fatalf("Error connecting to database: %v\n", err)
		}
	})
	return dbObj
}

// ConnectDb opens the database and returns a pointer to the bolt.DB instance
func ConnectDb() (*DBWrapper, error) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	basePath = GetBasePath()
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
	mongigoDbPath := basePath + "/monigo/monigo.db"
	if _, err := os.Stat(mongigoDbPath); err == nil {
		err = os.Remove(mongigoDbPath)
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

func GetServiceMetricsFromMonigoDbData() *models.ServiceMetrics {

	requestCount, totalDuration, memStats := core.GetServiceMetrics()
	serviceStat := core.GetProcessSats()

	var serviceMetrics models.ServiceMetrics

	serviceMetrics.Id = uuid.New()
	serviceMetrics.Load = fmt.Sprintf("%.2f", serviceStat.ProcCPUPercent)
	serviceMetrics.Cores = fmt.Sprintf("%.2f", serviceStat.ProcessUsedCores) + "PC / " +
		fmt.Sprintf("%.2f", serviceStat.SystemUsedCores) + "SC / " +
		strconv.Itoa(serviceStat.TotalLogicalCores) + "LC / " +
		strconv.Itoa(serviceStat.TotalCores) + "C"
	serviceMetrics.MemoryUsed = fmt.Sprintf("%.2f", serviceStat.ProcMemPercent)
	// serviceMetrics.UpTime = time.Since(serviceInfo.ServiceStartTime)
	serviceMetrics.NumberOfReqServerd = requestCount
	serviceMetrics.TotalDurationTookByAPI = totalDuration
	serviceMetrics.GoRoutines = int64(runtime.NumGoroutine())
	serviceMetrics.TotalAlloc = memStats.TotalAlloc
	serviceMetrics.MemoryAllocSys = memStats.Sys
	serviceMetrics.HeapAlloc = memStats.HeapAlloc
	serviceMetrics.HeapAllocSys = memStats.HeapSys
	serviceMetrics.TimeStamp = time.Now()

	return &serviceMetrics
}

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
