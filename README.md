<!-- ### Status: Testing going on for v1 release 🚀 -->
<p align="center">
  <img src="./static/assets/monigo-icon.png" width="200" title="Monigo Icon" alt="monigo-icon"/>
</p>

# MoniGo - Performance Monitoring for Go Applications

[![Go Report Card](https://goreportcard.com/badge/github.com/iyashjayesh/monigo)](https://goreportcard.com/report/github.com/iyashjayesh/monigo)
[![GoDoc](https://godoc.org/github.com/iyashjayesh/monigo?status.svg)](https://pkg.go.dev/github.com/iyashjayesh/monigo)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Visitors](https://api.visitorbadge.io/api/visitors?path=iyashjayesh%2Fmonigo%20&countColor=%23263759&style=flat)
![GitHub last commit](https://img.shields.io/github/last-commit/iyashjayesh/monigo)
<a href="https://www.producthunt.com/posts/monigo?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-monigo" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=487815&theme=light" alt="MoniGO - Go&#0032;App&#0032;Performance&#0032;Dashboard&#0032;in&#0032;10&#0032;Seconds&#0032;with&#0032;R&#0045;T&#0032;Insight&#0033; | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>

<!-- [![Github All Releases](https://img.shields.io/github/downloads/iyashjayesh/monigo/total.svg)](https://GitHub.com/iyashjayesh/monigo/releases/) -->

**MoniGo** is a performance monitoring library for Go applications. It provides real-time insights into application performance with an intuitive user interface, enabling developers to track and optimize both service-level and function-level metrics.

<div align="center" style="display: flex; flex-wrap: wrap; gap: 10px; border: 2px solid #ccc; padding: 10px;">
  <img src="./static/assets/ss/d1.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d10.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d2.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d7.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d8.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d3.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d4.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d5.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d6.png" alt="Dashboard" width="300">
  <img src="./static/assets/ss/d9.png" alt="Dashboard" width="300">
</div>

## Features

- **Real-Time Monitoring**: Access up-to-date performance metrics for your Go applications.
- **Detailed Insights**: Track and analyze both service and function-level performance.
- **Customizable Dashboard**: Manage performance data with an easy-to-use UI.
- **Visualizations**: Utilize graphs and charts to interpret performance trends.
- **Custom Thresholds**: Configure custom thresholds for your application's performance and resource usage.

## Installation

To install MoniGo, use the following command:

```bash
go get github.com/iyashjayesh/monigo@latest
```

## Example:

```go
package main

import (
    "github.com/iyashjayesh/monigo"
)

func main() {

	monigoInstance := &monigo.Monigo{
		ServiceName:             "data-api", // Mandatory field
		DashboardPort:           8080,       // Default is 8080
		DataPointsSyncFrequency: "5s",       // Default is 5 Minutes
		DataRetentionPeriod:     "4d",       // Default is 7 days. Supported values: "1h", "1d", "1w", "1m"
		TimeZone:                "Local",    // Default is Local timezone. Supported values: "Local", "UTC", "Asia/Kolkata", "America/New_York" etc. (https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)
		// MaxCPUUsage:             90,         // Default is 95%
		// MaxMemoryUsage:          90,         // Default is 95%
		// MaxGoRoutines:           100,        // Default is 100
	}

   	monigo.TraceFunction(highCPUUsage) // Trace function, when the function is called, it will be traced and the metrics will be displayed on the dashboard

	go monigoInstance.Start() // Starting monigo dashboard
	log.Println("Monigo dashboard started at port 8080")

  	// Optional
	// routinesStats := monigoInstance.GetGoRoutinesStats() // Get go routines stats
	// log.Println(routinesStats)

  	select {} // To keep the program running
}

// highCPUUsage is a function that simulates high CPU usage
func highCPUUsage() {
	// Simulate high CPU usage by performing heavy computations
	var sum float64
	for i := 0; i < 1e8; i++ { // 100 million iterations
		sum += math.Sqrt(float64(i))
	}
}
```

For more detailed usage instructions, refer to the documentation.
By default, the dashboard will be available at `http://localhost:8080/` else at the port you have provided.

### Note:

The `monigo.TraceFunction(func())` method accept `func(){}` as a type.

### Example Usage:

```go
func apiHandler(w http.ResponseWriter, r *http.Request) {
    // Trace function: when the highMemoryUsage function is called, it will be traced.
    monigo.TraceFunction(highMemoryUsage)
    w.Write([]byte("API1 response: memexpensiveFunc"))
}

func highMemoryUsage() {
    // Simulate high memory usage by allocating a large slice
    largeSlice := make([]float64, 1e8) // 100 million elements
    for i := 0; i < len(largeSlice); i++ {
        largeSlice[i] = float64(i)
    }
}
```

## Router Integration

MoniGo now supports integration with your existing HTTP server, allowing you to use your own router and authorization system. This is perfect for applications that need to integrate MoniGo as part of their existing infrastructure.

### Integration Options

#### 1. Full Integration (Recommended)
Register all MoniGo handlers (both API and static files) to your existing HTTP mux:

```go
package main

import (
    "log"
    "net/http"
    "github.com/iyashjayesh/monigo"
)

func main() {
    // Initialize MoniGo without starting the dashboard
    monigoInstance := &monigo.Monigo{
        ServiceName:             "my-service",
        DataPointsSyncFrequency: "5m",
        DataRetentionPeriod:     "7d",
        TimeZone:                "Local",
        CustomBaseAPIPath:       "/monitoring/api/v1", // Custom API path
    }

    // Initialize MoniGo (sets up metrics collection)
    monigoInstance.Initialize()

    // Create your own HTTP mux
    mux := http.NewServeMux()

    // Register all MoniGo handlers to your mux
    monigo.RegisterDashboardHandlers(mux, "/monitoring/api/v1")

    // Add your own routes
    mux.HandleFunc("/api/users", usersHandler)
    mux.HandleFunc("/health", healthHandler)

    log.Println("Server starting on :8080")
    log.Println("MoniGo dashboard: http://localhost:8080/")
    log.Println("MoniGo API: http://localhost:8080/monitoring/api/v1/")

    http.ListenAndServe(":8080", mux)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    // Trace functions for monitoring
    monigo.TraceFunction(func() {
        // Your function logic here
    })
    
    w.Write([]byte("Users endpoint"))
}
```

#### 2. API-Only Integration
Register only MoniGo API endpoints (useful when you want to handle static files yourself):

```go
// Register only API handlers
monigo.RegisterAPIHandlers(mux, "/monitoring/api/v1")

// Handle static files yourself
mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
```

#### 3. Static-Only Integration
Register only MoniGo static file handlers (useful when you want to handle API routing yourself):

```go
// Register only static handlers
monigo.RegisterStaticHandlers(mux)

// Handle API routing yourself
mux.HandleFunc("/api/metrics", customMetricsHandler)
```

#### 4. Maximum Flexibility
Get handlers as a map for integration with any HTTP router (Gin, Echo, etc.):

```go
// Get API handlers as a map
apiHandlers := monigo.GetAPIHandlers("/monitoring/api/v1")

// Get static handler
staticHandler := monigo.GetStaticHandler()

// Use with any router
for path, handler := range apiHandlers {
    router.Any(path, gin.WrapF(handler)) // Example with Gin
}
```

### Integration with Popular Frameworks

#### Gin Framework
```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/iyashjayesh/monigo"
)

func main() {
    monigoInstance := &monigo.Monigo{
        ServiceName: "gin-service",
        // ... other config
    }
    monigoInstance.Initialize()

    r := gin.Default()
    
    // Get and register MoniGo handlers
    apiHandlers := monigo.GetAPIHandlers("/monigo/api/v1")
    for path, handler := range apiHandlers {
        r.Any(path, gin.WrapF(handler))
    }
    
    staticHandler := monigo.GetStaticHandler()
    r.Any("/", gin.WrapF(staticHandler))

    r.Run(":8080")
}
```

#### Echo Framework
```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/iyashjayesh/monigo"
)

func main() {
    monigoInstance := &monigo.Monigo{
        ServiceName: "echo-service",
        // ... other config
    }
    monigoInstance.Initialize()

    e := echo.New()
    
    // Get and register MoniGo handlers
    apiHandlers := monigo.GetAPIHandlers("/monigo/api/v1")
    for path, handler := range apiHandlers {
        e.Any(path, echo.WrapHandler(http.HandlerFunc(handler)))
    }
    
    staticHandler := monigo.GetStaticHandler()
    e.Any("/", echo.WrapHandler(http.HandlerFunc(staticHandler)))

    e.Start(":8080")
}
```

### Available Integration Functions

| Function | Description |
|----------|-------------|
| `RegisterDashboardHandlers(mux, customPath)` | Register all handlers (API + static) |
| `RegisterAPIHandlers(mux, customPath)` | Register only API handlers |
| `RegisterStaticHandlers(mux)` | Register only static handlers |
| `GetAPIHandlers(customPath)` | Get API handlers as a map |
| `GetStaticHandler()` | Get static handler function |
| `Initialize()` | Initialize MoniGo without starting dashboard |

### Benefits of Router Integration

- **Unified Server**: Run MoniGo on the same port as your application
- **Custom Authorization**: Use your existing auth system to protect MoniGo endpoints
- **Custom Routing**: Integrate with your existing routing patterns
- **Framework Compatibility**: Works with any HTTP router (Gin, Echo, Chi, etc.)
- **Flexible Configuration**: Choose which parts of MoniGo to integrate

### Examples

Check out the complete examples in the `example/` directory:
- `example/router-integration/` - Standard HTTP mux integration
- `example/api-only-integration/` - API-only integration
- `example/gin-integration/` - Gin framework integration
- `example/echo-integration/` - Echo framework integration

## Bellow Reports are available

#### Note: You can download the reports in excel format.

1. **Load Statistics**: Provides an overview of the overall load of the service, CPU load, memory load, and system load.

| Field Name                | Value (Datatype) |
| ------------------------- | ---------------- |
| `overall_load_of_service` | `float64`        |
| `service_cpu_load`        | `float64`        |
| `service_memory_load`     | `float64`        |
| `system_cpu_load`         | `float64`        |
| `system_memory_load`      | `float64`        |

2. **CPU Statistics**: Displays the total number of cores, cores used by the service, and cores used by the system.

| Field Name              | Value (Datatype) |
| ----------------------- | ---------------- |
| `total_cores`           | `int`            |
| `cores_used_by_service` | `int`            |
| `cores_used_by_system`  | `int`            |

3. **Memory Statistics**: Shows the total system memory, memory used by the system, memory used by the service, available memory, GC pause duration, and stack memory usage.

| Field Name               | Value (Datatype) |
| ------------------------ | ---------------- |
| `total_system_memory`    | `float64`        |
| `memory_used_by_system`  | `float64`        |
| `memory_used_by_service` | `float64`        |
| `available_memory`       | `float64`        |
| `gc_pause_duration`      | `float64`        |
| `stack_memory_usage`     | `float64`        |

4. **Memory Profile**: Provides information on heap allocation by the service, heap allocation by the system, total allocation by the service, and total memory by the OS.

| Field Name               | Value (Datatype) |
| ------------------------ | ---------------- |
| `heap_alloc_by_service`  | `float64`        |
| `heap_alloc_by_system`   | `float64`        |
| `total_alloc_by_service` | `float64`        |
| `total_memory_by_os`     | `float64`        |

5. **Network IO**: Displays the number of bytes sent and received.

| Field Name       | Value (Datatype) |
| ---------------- | ---------------- |
| `bytes_sent`     | `float64`        |
| `bytes_received` | `float64`        |

6. **Health Metrics**: Provides an overall health percentage for the service.

| Field Name               | Value (Datatype) |
| ------------------------ | ---------------- |
| `service_health_percent` | `float64`        |
| `system_health_percent`  | `float64`        |

## API Reference

- You can access the MoniGo API by visiting the following URL: http://localhost:8080/monigo/api/v1/<endpoint> (replace `<endpoint>` with the desired endpoint).
- **Note**: When using router integration, the API path can be customized using the `CustomBaseAPIPath` field or by passing a custom path to the registration functions.
- API endpoints are available for the following:

| Endpoint                           | Description           | Method | Request                                               | Response | Example                                            |
| ---------------------------------- | --------------------- | ------ | ----------------------------------------------------- | -------- | -------------------------------------------------- |
| `/monigo/api/v1/metrics`           | Get all metrics       | GET    | None                                                  | JSON     | [Example](./static/API/Res/metrics.json)           |
| `/monigo/api/v1/go-routines-stats` | Get go routines stats | GET    | None                                                  | JSON     | [Example](./static/API/Res/go-routines-stats.json) |
| `/monigo/api/v1/service-info`      | Get service info      | GET    | None                                                  | JSON     | [Example](./static/API/Res/service-info.json)      |
| `/monigo/api/v1/service-metrics`   | Get service metrics   | POST   | JSON [Example](./static/API/Req/service-metrics.json) | JSON     | [Example](./static/API/Res/service-metrics.json)   |
| `/monigo/api/v1/reports`           | Get history data      | POST   | JSON [Example](./static/API/Req/reports.json)         | JSON     | [Example](./static/API/Res/reports.json)           |

## Contributing

We welcome contributions! If you encounter any issues or have suggestions, please submit a pull request or open an issue.

**If you find MoniGo useful, consider giving it a star! ⭐**

## Contact

For questions or feedback, please open an issue or contact me at `iyashjayesh@gmail.com` or at [LinkedIn](https://www.linkedin.com/in/iyashjayesh/)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=iyashjayesh/monigo&type=Date)](https://star-history.com/#iyashjayesh/monigo&Date)

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE file](https://github.com/iyashjayesh/monigo?tab=Apache-2.0-1-ov-file) for details.
