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
<!-- <a href="https://www.producthunt.com/posts/monigo?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-monigo" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=487815&theme=light" alt="MoniGO - Go&#0032;App&#0032;Performance&#0032;Dashboard&#0032;in&#0032;10&#0032;Seconds&#0032;with&#0032;R&#0045;T&#0032;Insight&#0033; | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a> -->

<!-- [![Github All Releases](https://img.shields.io/github/downloads/iyashjayesh/monigo/total.svg)](https://GitHub.com/iyashjayesh/monigo/releases/) -->

**MoniGo** is a performance monitoring library for Go applications. It provides real-time insights into application performance with an intuitive user interface, enabling developers to track and optimize both service-level and function-level metrics.

<!-- <div align="center" style="display: flex; flex-wrap: wrap; gap: 10px; border: 2px solid #ccc; padding: 10px;">
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
</div> -->

<div align="center">
  <img src="monigo.gif" width="100%" alt="monigo-gif">
</div>

## Features

- **Real-Time Monitoring**: Access up-to-date performance metrics for your Go applications.
- **Detailed Insights**: Track and analyze both service and function-level performance.
- **Disk I/O Monitoring**: Monitor disk read/write bytes and system disk load.
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

func main() {
    // New way: Use Builder Pattern for clean initialization
    monigoInstance := monigo.NewBuilder().
        WithServiceName("data-api").
        WithPort(8080).
        WithRetentionPeriod("4d").
        WithDataPointsSyncFrequency("5s").
        Build()

   	monigo.TraceFunction(highCPUUsage) // Trace function

	go func() {
        // Start returns an error now, so handle it!
        if err := monigoInstance.Start(); err != nil {
            log.Fatalf("Failed to start MoniGo: %v", err)
        }
    }()
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

## Function Tracing

MoniGo provides powerful function tracing capabilities to monitor the performance of your application functions. You can trace functions with any signature, including those with parameters and return values.

### Available Tracing Methods

#### 1. Legacy Method (Backward Compatible)
The original `TraceFunction` method for functions without parameters:

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

#### 2. Enhanced Method - Functions with Parameters
Use `TraceFunctionWithArgs` to trace functions that take parameters:

```go
// Business logic function with parameters
func processUser(userID string, userName string) {
    // Your business logic here
    time.Sleep(100 * time.Millisecond)
    _ = make([]byte, 1024*1024) // 1MB allocation
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    userName := r.URL.Query().Get("name")
    
    // NEW WAY: Direct function tracing with parameters
    monigo.TraceFunctionWithArgs(processUser, userID, userName)
    
    w.Write([]byte("User processed"))
}
```

#### 3. Enhanced Method - Functions with Return Values
Use `TraceFunctionWithReturn` to trace functions that return values:

```go
// Business logic function with return value
func calculateTotal(items []Item) float64 {
    var total float64
    for _, item := range items {
        total += item.Price
    }
    return total
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
    items := []Item{
        {Name: "Laptop", Price: 999.99},
        {Name: "Mouse", Price: 29.99},
    }
    
    // NEW WAY: Trace function with return value
    total := monigo.TraceFunctionWithReturn(calculateTotal, items).(float64)
    
    w.Write([]byte(fmt.Sprintf("Total: $%.2f", total)))
}
```

#### 4. Enhanced Method - Functions with Multiple Return Values
Use `TraceFunctionWithReturns` to trace functions that return multiple values:

```go
// Business logic function with multiple return values
func processData(data string) (Result, error) {
    // Processing logic
    if data == "error" {
        return Result{}, fmt.Errorf("processing error")
    }
    return Result{Success: true, Message: "OK"}, nil
}

func processHandler(w http.ResponseWriter, r *http.Request) {
    data := "test-data"
    
    // NEW WAY: Trace function with multiple returns
    results := monigo.TraceFunctionWithReturns(processData, data)
    
    if len(results) >= 2 {
        result := results[0].(Result)
        err := results[1].(error)
        
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte(fmt.Sprintf("Error: %v", err)))
            return
        }
        
        w.Write([]byte(fmt.Sprintf("Result: %+v", result)))
    }
}
```

**Alternative: Using the first return value only**
```go
// For functions with multiple returns, you can still use TraceFunctionWithReturn
// to get just the first return value
result := monigo.TraceFunctionWithReturn(processData, data).(Result)
// Note: This ignores the error return value
```

### Benefits of Enhanced Tracing

- **Cleaner Code**: No need to wrap functions in anonymous functions
- **Better Function Identification**: Actual function names appear in metrics instead of anonymous functions
- **Type Safety**: Compile-time checking of function signatures
- **Flexibility**: Support for any function signature (parameters, return values, etc.)
- **Backward Compatibility**: Existing code continues to work without changes

### Function Name Generation

The enhanced tracing methods automatically generate descriptive function names that include:
- Function name
- Parameter types: `functionName(string,int)`
- Return types: `functionName(string,int)->(float64,error)`

This makes it easier to identify and analyze specific function calls in the dashboard.

### Handling Multiple Return Values

When dealing with functions that return multiple values, you have several options:

#### Option 1: Get All Return Values (Recommended)
```go
func processData(data string) (Result, error) {
    // Your logic here
    return result, nil
}

// Get all return values
results := monigo.TraceFunctionWithReturns(processData, data)
result := results[0].(Result)
err := results[1].(error)
```

#### Option 2: Get Only the First Return Value
```go
// Get only the first return value (ignores error)
result := monigo.TraceFunctionWithReturn(processData, data).(Result)
```

#### Option 3: Handle Different Return Counts
```go
results := monigo.TraceFunctionWithReturns(myFunction, args...)

switch len(results) {
case 0:
    // Function returns nothing
case 1:
    // Function returns one value
    value := results[0]
case 2:
    // Function returns two values (common pattern: result, error)
    value := results[0]
    err := results[1].(error)
case 3:
    // Function returns three values
    value1 := results[0]
    value2 := results[1]
    err := results[2].(error)
default:
    // Function returns many values
    // Handle accordingly
}
```

### Example: Complete Usage

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/iyashjayesh/monigo"
)

func main() {
    monigoInstance := &monigo.Monigo{
        ServiceName: "my-service",
        DashboardPort: 8080,
    }
    
    go monigoInstance.Start()
    
    http.HandleFunc("/api/user", userHandler)
    http.HandleFunc("/api/calculate", calculateHandler)
    http.ListenAndServe(":8000", nil)
}

// Functions with different signatures
func processUser(userID string, userName string) {
    // Business logic
}

func calculateTotal(items []Item) float64 {
    // Calculation logic
    return 0.0
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    // Enhanced tracing - direct function calls
    monigo.TraceFunctionWithArgs(processUser, "123", "John")
    w.Write([]byte("User processed"))
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
    items := []Item{{Name: "Item1", Price: 10.0}}
    
    // Enhanced tracing - with return value
    total := monigo.TraceFunctionWithReturn(calculateTotal, items).(float64)
    w.Write([]byte(fmt.Sprintf("Total: %.2f", total)))
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
    monigoInstance := monigo.NewBuilder().
        WithServiceName("my-service").
        WithDataPointsSyncFrequency("5m").
        WithRetentionPeriod("7d").
        WithTimeZone("Local").
        WithCustomBaseAPIPath("/monitoring/api/v1").
        Build()

    // Initialize MoniGo (sets up metrics collection)
    if err := monigoInstance.Initialize(); err != nil {
        log.Fatalf("Failed to initialize MoniGo: %v", err)
    }

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

## Dashboard Security

MoniGo now includes comprehensive security features to protect your dashboard and API endpoints in production environments. You can use built-in middleware or implement custom authentication to secure access to your monitoring data.

### Security Features

- **Built-in Middleware**: Pre-built security middleware for common use cases
- **Custom Authentication**: Support for custom authentication functions
- **Middleware Chains**: Chain multiple middleware for layered security
- **Rate Limiting**: Built-in rate limiting to prevent abuse
- **IP Restrictions**: IP whitelisting and blacklisting support
- **Request Logging**: Comprehensive request logging for audit trails

### Built-in Security Middleware

MoniGo provides several built-in security middleware functions:

#### Basic Authentication
```go
monigo.BasicAuthMiddleware("username", "password")
```

#### API Key Authentication
```go
monigo.APIKeyMiddleware("your-secret-api-key")
```

#### IP Whitelist
```go
monigo.IPWhitelistMiddleware([]string{"127.0.0.1", "192.168.1.0/24"})
```

#### Rate Limiting
```go
monigo.RateLimitMiddleware(100, time.Minute) // 100 requests per minute
```

#### Request Logging
```go
monigo.LoggingMiddleware()
```

### Advanced Security Features

#### Static File Handling
MoniGo automatically bypasses authentication for static files (CSS, JS, images, etc.) to ensure the dashboard UI loads correctly:

```go
// Static files are automatically excluded from authentication
// This includes: .css, .js, .png, .jpg, .gif, .svg, .ico, .woff, .woff2, etc.
// And paths: /css/, /js/, /assets/, /images/, /fonts/, /static/
```

#### IP Whitelist with Debug Logging
The IP whitelist middleware includes comprehensive debug logging to help troubleshoot access issues:

```go
// IP whitelist with debug logging
monigo.IPWhitelistMiddleware([]string{
    "127.0.0.1",      // IPv4 localhost
    "::1",            // IPv6 localhost
    "192.168.1.0/24", // Local network
    "10.0.0.0/8",     // Private network
})
```

#### Router-Specific Handlers
MoniGo provides framework-specific handlers for seamless integration:

```go
// Gin Framework
ginHandler := monigo.GetGinHandler("/monigo/api/v1")

// Echo Framework  
echoHandler := monigo.GetEchoHandler("/monigo/api/v1")

// Fiber Framework
fiberHandler := monigo.GetFiberHandler("/monigo/api/v1")
```

### Security Configuration

#### Using Built-in Middleware

```go
monigoInstance := monigo.NewBuilder().
    WithServiceName("my-service").
    
    // Dashboard security (for static files)
    WithDashboardMiddleware(
        monigo.BasicAuthMiddleware("admin", "password"),
        monigo.LoggingMiddleware(),
    ).
    
    // API security (for API endpoints)
    WithAPIMiddleware(
        monigo.APIKeyMiddleware("api-key"),
        monigo.RateLimitMiddleware(100, time.Minute),
    ).
    Build()
```

#### Using Custom Authentication

```go
monigoInstance := monigo.NewBuilder().
    WithServiceName("my-service").
    
    // Custom authentication function
    WithAuthFunction(func(r *http.Request) bool {
        return r.Header.Get("X-API-Key") == "secret-key"
    }).
    Build()
```

### Secured Handler Functions

MoniGo provides secured versions of all handler functions:

| Function | Description |
|----------|-------------|
| `GetSecuredUnifiedHandler(m, customPath)` | Get unified handler with middleware |
| `GetSecuredAPIHandlers(m, customPath)` | Get API handlers with middleware |
| `GetSecuredStaticHandler(m)` | Get static handler with middleware |
| `StartSecuredDashboard(m)` | Start dashboard with middleware |
| `RegisterSecuredDashboardHandlers(mux, m, customPath)` | Register secured dashboard handlers |
| `RegisterSecuredAPIHandlers(mux, m, customPath)` | Register secured API handlers |
| `RegisterSecuredStaticHandlers(mux, m)` | Register secured static handlers |

### Security Examples

Check out the comprehensive security examples in the `example/security-examples/` directory:

#### Core Security Examples
- **Basic Authentication** (`basic-auth/`) - HTTP Basic Auth with rate limiting
- **API Key Authentication** (`api-key/`) - API key via header/query parameter  
- **IP Whitelist** (`ip-whitelist-example/`) - IP-based access control with debug logging
- **Custom Authentication** (`custom-auth/`) - Custom auth function with headers/query params

#### Router Integration Examples
- **Gin Integration** (`gin/`) - Gin framework with security middleware
- **Echo Integration** (`echo/`) - Echo framework with security middleware
- **Fiber Integration** (`fiber/`) - Fiber framework with security middleware
- **Chi Integration** (`chi/`) - Chi router with security middleware

#### Running Security Examples

Each example has its own `go.mod` file and can be run independently:

```bash
# Basic Authentication Example
cd example/security-examples/basic-auth
go run .

# API Key Example  
cd example/security-examples/api-key
go run .

# IP Whitelist Example
cd example/security-examples/ip-whitelist-example
go run .

# Custom Authentication Example
cd example/security-examples/custom-auth
go run .

# Router Integration Examples
cd example/security-examples/gin
go run .

cd example/security-examples/echo
go run .

cd example/security-examples/fiber
go run .

cd example/security-examples/chi
go run .
```

#### Example Access URLs

- **Basic Auth**: `http://localhost:8080/` (username: `admin`, password: `monigo-secure-2024`)
- **API Key**: `http://localhost:8080/?api_key=monigo-secret-key-2024`
- **IP Whitelist**: `http://localhost:8080/` (localhost only)
- **Custom Auth**: `http://localhost:8080/?secret=monigo-admin-secret` or with headers
- **Router Examples**: Same URLs as above with framework-specific implementations

### JavaScript Authentication Integration

MoniGo's dashboard JavaScript automatically handles authentication for different security methods:

#### API Key Authentication
The dashboard automatically detects API keys from URL parameters and includes them in all API requests:

```javascript
// Automatically extracts api_key from URL and includes in requests
// URL: http://localhost:8080/?api_key=your-secret-key
// All API calls will include: ?api_key=your-secret-key
```

#### Basic Authentication
For basic auth, the browser handles credentials automatically when prompted:

```javascript
// Browser automatically includes Authorization header
// No additional JavaScript configuration needed
```

#### Custom Authentication
Supports custom headers and query parameters:

```javascript
// Automatically adds custom headers and query parameters
// Based on URL parameters or predefined authentication logic
```

### Production Security Best Practices

1. **Use Strong Credentials**: Always use strong, unique passwords and API keys
2. **Enable HTTPS**: Always use HTTPS in production environments
3. **Implement Rate Limiting**: Use rate limiting to prevent abuse
4. **IP Restrictions**: Use IP whitelisting for internal networks
5. **Request Logging**: Enable logging to monitor access patterns
6. **Regular Rotation**: Regularly rotate API keys and passwords
7. **Environment Variables**: Store credentials in environment variables
8. **Monitor Access**: Set up monitoring and alerting for security events
9. **Static File Security**: Static files (CSS, JS, images) bypass authentication automatically
10. **Debug Logging**: Use debug logging to monitor authentication attempts
11. **Router Compatibility**: Test security middleware with your chosen HTTP framework
12. **JavaScript Integration**: Ensure dashboard JavaScript handles your authentication method

### Security Troubleshooting

#### Common Issues and Solutions

**Issue: Dashboard loads but CSS/JS files show 401 Unauthorized**
- **Solution**: Static files should bypass authentication automatically. Check that `isStaticFile()` function is working correctly.

**Issue: IP Whitelist blocking localhost access**
- **Solution**: Add both IPv4 (`127.0.0.1`) and IPv6 (`::1`) localhost addresses to your whitelist.

**Issue: API calls failing with authentication errors**
- **Solution**: Ensure JavaScript is including authentication credentials. Check browser network tab for request headers/parameters.

**Issue: Router integration not working**
- **Solution**: Use the appropriate framework-specific handler (`GetGinHandler`, `GetEchoHandler`, `GetFiberHandler`).

**Issue: Rate limiting too restrictive**
- **Solution**: Adjust rate limit parameters: `RateLimitMiddleware(requests, timeWindow)`.

#### Debug Mode
Enable debug logging to troubleshoot authentication issues:

```go
// Add debug middleware to see what's happening
DashboardMiddleware: []func(http.Handler) http.Handler{
    debugMiddleware(), // Your custom debug middleware
    monigo.LoggingMiddleware(),
    // ... other middleware
}
```

### Quick Start with Security

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/iyashjayesh/monigo"
)

func main() {
    // Initialize MoniGo with security
    monigoInstance := &monigo.Monigo{
        ServiceName: "secure-service",
        DashboardMiddleware: []func(http.Handler) http.Handler{
            monigo.BasicAuthMiddleware("admin", "secure-password"),
            monigo.LoggingMiddleware(),
        },
        APIMiddleware: []func(http.Handler) http.Handler{
            monigo.RateLimitMiddleware(100, time.Minute),
        },
    }
    
    // Initialize and start secured dashboard
    monigoInstance.Initialize()
    
    if err := monigo.StartSecuredDashboard(monigoInstance); err != nil {
        log.Fatal("Failed to start secured dashboard:", err)
    }
}
```

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
