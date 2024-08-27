// Functions for serving the dashboard and handling HTTP requests.
package monigo

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	//go:embed static/*
	staticFiles      embed.FS
	serviceStartTime time.Time = time.Now()
	Once             sync.Once = sync.Once{}
	Db               *bolt.DB
	basePath         string
	serviceInfo      ServiceInfo
	dbObj            *DBWrapper
)

func init() {

	Once.Do(func() {
		basePath = getBasePath()

		// Connect to the database
		var err error
		dbObj, err = connectDb()
		if err != nil {
			log.Fatalf("Error connecting to database: %v\n", err)
		}
	})
}

func StartDashboard(addr, serviceName string) {

	serviceInfo.ServiceName = serviceName
	serviceInfo.ServiceStartTime = serviceStartTime
	serviceInfo.GoVerison = runtime.Version()
	serviceInfo.TimeStamp = serviceStartTime

	dbObj.StoreServiceInfo(&serviceInfo)

	serviceInfo, err := dbObj.GetServiceInfo(serviceInfo.ServiceName)
	if err != nil {
		log.Fatalf("Error getting service info: %v\n", err)
	}

	log.Printf("Service Name: %s\nService Start Time: %s\nGo Version: %s\nTime Stamp: %s\n", serviceInfo.ServiceName, serviceInfo.ServiceStartTime, serviceInfo.GoVerison, serviceInfo.TimeStamp)

	// Serve the index.html at the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			file, err := staticFiles.ReadFile("static/index.html")
			if err != nil {
				http.Error(w, "Could not load index.html", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(file)
			return
		}

		// Serve other static files (CSS, JS, etc.) correctly
		http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))).ServeHTTP(w, r)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		unit := r.URL.Query().Get("unit")
		if unit == "" {
			unit = "MB" // Default unit
		}

		requestCount, totalDuration, memStats := GetServiceMetrics()
		serviceStat := GetProcessSats()

		// Convert bytes to different units
		bytesToUnit := func(bytes uint64) float64 {
			switch unit {
			case "KB":
				return float64(bytes) / 1024.0
			case "MB":
				return float64(bytes) / 1048576.0
			default: // "bytes"
				return float64(bytes)
			}
		}

		SystemUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.SystemUsedCores)
		ProcessUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.ProcessUsedCores)

		core := ProcessUsedCoresToString + "PC / " +
			SystemUsedCoresToString + "SC / " +
			strconv.Itoa(serviceStat.TotalLogicalCores) + "LC / " +
			strconv.Itoa(serviceStat.TotalCores) + "C"

		// ProcMemPercent
		memoryUsed := fmt.Sprintf("%.2f", serviceStat.ProcMemPercent)

		metrics := fmt.Sprintf(
			"Service Name: %s\nService Start Time: %s\nGoroutines: %d\nRequests: %d\nTotal Duration: %s\n\nMemory Usage (%s):\nAlloc: %.2f %s\nTotalAlloc: %.2f %s\nSys: %.2f %s\nHeapAlloc: %.2f %s\nHeapSys: %.2f %s\nGo Version: %s\n Load: %s\nCores: %s\n Memory Used: %s\n",
			serviceName,
			serviceStartTime.Format(time.RFC3339),
			GetGoroutineCount(),
			requestCount,
			totalDuration,
			unit,
			bytesToUnit(memStats.Alloc),
			unit,
			bytesToUnit(memStats.TotalAlloc),
			unit,
			bytesToUnit(memStats.Sys),
			unit,
			bytesToUnit(memStats.HeapAlloc),
			unit,
			bytesToUnit(memStats.HeapSys),
			unit,
			runtime.Version(),
			fmt.Sprintf("%.2f", serviceStat.ProcCPUPercent),
			core,
			memoryUsed,
		)

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(metrics))
	})

	http.HandleFunc("/function-metrics", func(w http.ResponseWriter, r *http.Request) {
		unit := r.URL.Query().Get("unit")
		if unit == "" {
			unit = "MB" // Default unit
		}

		// Convert bytes to different units
		bytesToUnit := func(bytes uint64) float64 {
			switch unit {
			case "KB":
				return float64(bytes) / 1024.0
			case "MB":
				return float64(bytes) / 1048576.0
			default: // "bytes"
				return float64(bytes)
			}
		}

		var results string
		mu.Lock()
		for name, metrics := range functionMetrics {
			results += fmt.Sprintf(
				"Function: %s\nFunction Ran At: %s\nCPU Profile: %s\nExecution Time: %s\nMemory Usage: %.2f %s\nGoroutines: %d\n\n",
				name,
				metrics.FunctionLastRanAt.Format(time.RFC3339),
				metrics.CPUProfile,
				metrics.ExecutionTime,
				bytesToUnit(metrics.MemoryUsage),
				unit,
				metrics.GoroutineCount,
			)
		}
		mu.Unlock()

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(results))
	})

	http.HandleFunc("/cpu-metrics", profileHandler)
	http.HandleFunc("/mem-metrics", profileHandler)

	// fs := http.FileServer(http.FS(staticFiles))
	// http.Handle("/static/", http.StripPrefix("/static/", fs)) // Serve all other static files

	fmt.Printf("Starting dashboard on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error starting dashboard: %v\n", err)
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Generating profile\n")
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	profilesFolderPath := fmt.Sprintf("%s/profiles", basePath)

	cmd := exec.Command("go", "tool", "pprof", "-svg", profilesFolderPath)
	output, err := cmd.Output()
	if err != nil {
		errMsg := fmt.Sprintf("failed to generate profile, given path %s, error: %v", profilesFolderPath, err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	if _, err := w.Write(output); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
