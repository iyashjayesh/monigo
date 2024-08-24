// Function for recording and retrieving metrics.
package monigo

import (
	"runtime"
	"sync"
	"time"
)

var (
	mu              sync.Mutex
	requestCount    int
	totalDuration   time.Duration
	functionMetrics = make(map[string]*FunctionMetrics)
)

type FunctionMetrics struct {
	FunctionLastRanAt time.Time
	CPUProfile        string
	MemoryUsage       uint64
	GoroutineCount    int
	ExecutionTime     time.Duration
}

func RecordRequestDuration(duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	requestCount++
	totalDuration += duration
}

func GetServiceMetrics() (int, time.Duration, *runtime.MemStats) {
	mu.Lock()
	defer mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return requestCount, totalDuration, &memStats
}

func GetFunctionMetrics(functionName string) *FunctionMetrics {
	mu.Lock()
	defer mu.Unlock()
	return functionMetrics[functionName]
}
