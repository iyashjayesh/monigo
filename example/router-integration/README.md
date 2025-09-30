# MoniGo Router Integration Examples

This directory contains examples showing how to integrate MoniGo with different Go HTTP routers and frameworks. MoniGo provides flexible integration options that allow you to embed the monitoring dashboard and API into your existing application without running a separate server.

## Quick Start

Choose your preferred framework and follow the corresponding example:

- **[Standard HTTP Mux](#standard-http-mux)** - Basic `net/http` integration
- **[Gin Framework](#gin-framework)** - High-performance HTTP web framework
- **[Echo Framework](#echo-framework)** - High-performance, minimalist web framework
- **[Fiber Framework](#fiber-framework)** - Express.js inspired web framework
- **[Gorilla Mux](#gorilla-mux)** - Powerful HTTP router and URL matcher

## Prerequisites

- Go 1.19 or higher
- Basic understanding of Go HTTP routing

## Integration Methods

MoniGo provides several integration methods:

### 1. Unified Handler (Recommended)
```go
// Get a single handler that manages both API and static files
unifiedHandler := monigo.GetUnifiedHandler("/monigo/api/v1")
```

### 2. Separate Handlers
```go
// Get only API handlers
apiHandlers := monigo.GetAPIHandlers("/monigo/api/v1")

// Get only static file handler
staticHandler := monigo.GetStaticHandler()
```

### 3. Fiber-Specific Handler
```go
// Get a Fiber-compatible handler
fiberHandler := monigo.GetFiberHandler("/monigo/api/v1")
```

## Examples

### Standard HTTP Mux

**File:** `standard-mux-integration/main.go`

```go
package main

import (
    "log"
    "net/http"
    "github.com/iyashjayesh/monigo"
)

func main() {
    // Initialize MoniGo
    monigoInstance := &monigo.Monigo{
        ServiceName:             "standard-mux-example",
        DataPointsSyncFrequency: "5m",
        DataRetentionPeriod:     "7d",
        TimeZone:                "Local",
        CustomBaseAPIPath:       "/monigo/api/v1",
    }
    monigoInstance.Initialize()

    // Create HTTP mux
    mux := http.NewServeMux()
    
    // Add your routes
    mux.HandleFunc("/api/users", usersHandler)
    mux.HandleFunc("/health", healthHandler)
    
    // Add MoniGo unified handler
    unifiedHandler := monigo.GetUnifiedHandler("/monigo/api/v1")
    mux.HandleFunc("/", unifiedHandler)
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

**Run:**
```bash
cd standard-mux-integration
go run main.go
```

### Gin Framework

**File:** `gin-integration/main.go`

```go
package main

import (
    "log"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/iyashjayesh/monigo"
)

func main() {
    // Initialize MoniGo
    monigoInstance := &monigo.Monigo{
        ServiceName:             "gin-integration-example",
        DataPointsSyncFrequency: "5m",
        DataRetentionPeriod:     "7d",
        TimeZone:                "Local",
        CustomBaseAPIPath:       "/monigo/api/v1",
    }
    monigoInstance.Initialize()

    // Create Gin router
    r := gin.Default()
    
    // Add your routes
    r.GET("/api/users", usersHandler)
    r.GET("/health", healthHandler)
    
    // Add MoniGo handlers
    unifiedHandler := monigo.GetUnifiedHandler("/monigo/api/v1")
    r.GET("/", func(c *gin.Context) {
        unifiedHandler(c.Writer, c.Request)
    })
    r.GET("/css/*filepath", func(c *gin.Context) {
        unifiedHandler(c.Writer, c.Request)
    })
    r.GET("/js/*filepath", func(c *gin.Context) {
        unifiedHandler(c.Writer, c.Request)
    })
    r.GET("/assets/*filepath", func(c *gin.Context) {
        unifiedHandler(c.Writer, c.Request)
    })
    r.GET("/monigo/*filepath", func(c *gin.Context) {
        unifiedHandler(c.Writer, c.Request)
    })
    r.POST("/monigo/*filepath", func(c *gin.Context) {
        unifiedHandler(c.Writer, c.Request)
    })
    
    log.Println("Server starting on :8080")
    log.Fatal(r.Run(":8080"))
}
```

**Run:**
```bash
cd gin-integration
go run main.go
```

### Echo Framework

**File:** `echo-integration/main.go`

```go
package main

import (
    "log"
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/iyashjayesh/monigo"
)

func main() {
    // Initialize MoniGo
    monigoInstance := &monigo.Monigo{
        ServiceName:             "echo-integration-example",
        DataPointsSyncFrequency: "5m",
        DataRetentionPeriod:     "7d",
        TimeZone:                "Local",
        CustomBaseAPIPath:       "/monigo/api/v1",
    }
    monigoInstance.Initialize()

    // Create Echo instance
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    
    // Add your routes
    e.GET("/api/users", usersHandler)
    e.GET("/health", healthHandler)
    
    // Add MoniGo unified handler
    unifiedHandler := monigo.GetUnifiedHandler("/monigo/api/v1")
    e.Any("/*", echo.WrapHandler(http.HandlerFunc(unifiedHandler)))
    
    log.Println("Server starting on :8080")
    log.Fatal(e.Start(":8080"))
}
```

**Run:**
```bash
cd echo-integration
go run main.go
```

### Fiber Framework

**File:** `fiber-integration/main.go`

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/iyashjayesh/monigo"
)

func main() {
    // Initialize MoniGo
    monigoInstance := &monigo.Monigo{
        ServiceName:             "fiber-integration-example",
        DataPointsSyncFrequency: "5m",
        DataRetentionPeriod:     "7d",
        TimeZone:                "Local",
        CustomBaseAPIPath:       "/monigo/api/v1",
    }
    monigoInstance.Initialize()

    // Create Fiber app
    app := fiber.New()
    app.Use(logger.New())
    app.Use(recover.New())
    
    // Add your routes
    app.Get("/api/users", usersHandler)
    app.Get("/health", healthHandler)
    
    // Add MoniGo Fiber handler
    fiberHandler := monigo.GetFiberHandler("/monigo/api/v1")
    app.All("*", fiberHandler)
    
    log.Println("Server starting on :8080")
    log.Fatal(app.Listen(":8080"))
}
```

**Run:**
```bash
cd fiber-integration
go run main.go
```

### Gorilla Mux

**File:** `gorilla-mux-integration/main.go`

```go
package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/iyashjayesh/monigo"
)

func main() {
    // Initialize MoniGo
    monigoInstance := &monigo.Monigo{
        ServiceName:             "gorilla-mux-example",
        DataPointsSyncFrequency: "5m",
        DataRetentionPeriod:     "7d",
        TimeZone:                "Local",
        CustomBaseAPIPath:       "/monigo/api/v1",
    }
    monigoInstance.Initialize()

    // Create Gorilla Mux router
    r := mux.NewRouter()
    
    // Add your routes
    r.HandleFunc("/api/users", usersHandler).Methods("GET")
    r.HandleFunc("/health", healthHandler).Methods("GET")
    
    // Add MoniGo unified handler
    unifiedHandler := monigo.GetUnifiedHandler("/monigo/api/v1")
    r.PathPrefix("/").HandlerFunc(unifiedHandler)
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

**Run:**
```bash
cd gorilla-mux-integration
go run main.go
```

## Available Endpoints

After integration, the following endpoints will be available:

### Dashboard
- `GET /` - MoniGo dashboard (HTML interface)

### API Endpoints
- `GET /monigo/api/v1/service-info` - Service information
- `GET /monigo/api/v1/metrics` - Service metrics
- `POST /monigo/api/v1/service-metrics` - Historical metrics data
- `GET /monigo/api/v1/go-routines-stats` - Go routines statistics
- `GET /monigo/api/v1/function` - Function trace details
- `GET /monigo/api/v1/function-details` - Function metrics
- `GET /monigo/api/v1/reports` - Report data

### Static Assets
- `GET /css/*` - CSS files
- `GET /js/*` - JavaScript files
- `GET /assets/*` - Images and other assets

## Configuration Options

### MoniGo Configuration
```go
monigoInstance := &monigo.Monigo{
    ServiceName:             "your-service-name",     // Required
    DataPointsSyncFrequency: "5m",                    // Default: "5m"
    DataRetentionPeriod:     "7d",                    // Default: "7d"
    TimeZone:                "Local",                 // Default: "Local"
    CustomBaseAPIPath:       "/monigo/api/v1",        // Custom API path
    MaxCPUUsage:            95.0,                     // Default: 95%
    MaxMemoryUsage:         95.0,                     // Default: 95%
    MaxGoRoutines:          100,                      // Default: 100
}
```

### Custom API Path
You can customize the API path to avoid conflicts:
```go
// Use custom path
unifiedHandler := monigo.GetUnifiedHandler("/custom/api/path")

// Your API endpoints will be available at:
// GET /custom/api/path/service-info
// GET /custom/api/path/metrics
// etc.
```

## Function Tracing

MoniGo supports function tracing to monitor performance:

```go
func yourHandler(c *gin.Context) {
    // Trace this function for monitoring
    monigo.TraceFunction(func() {
        // Your business logic here
        time.Sleep(100 * time.Millisecond)
    })
    
    c.JSON(200, gin.H{"message": "success"})
}
```

## Known Issues

### Gin Framework
- **Status Code Issue**: Static files and API responses may return HTTP 404 status codes even though content is served correctly. This is a known limitation with Gin's response handling.

### Workarounds
1. **For Production**: Consider using Echo or Gorilla Mux for better compatibility
2. **For Development**: The dashboard and API functionality work correctly despite the status code issue
3. **Alternative**: Use MoniGo's standalone mode with `Start()` instead of router integration

## Testing

Test your integration:

```bash
# Test dashboard
curl http://localhost:8080/

# Test API
curl http://localhost:8080/monigo/api/v1/service-info

# Test custom API
curl http://localhost:8080/api/users

# Test static files
curl http://localhost:8080/css/monigo-styles.css
```

## Additional Resources

- [MoniGo Documentation](../README.md)
- [Gin Framework](https://gin-gonic.com/)
- [Echo Framework](https://echo.labstack.com/)
- [Fiber Framework](https://gofiber.io/)
- [Gorilla Mux](https://github.com/gorilla/mux)

## Contributing

Found an issue or want to improve the integration examples? Please:

1. Check existing issues
2. Create a new issue with details
3. Submit a pull request with your improvements

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.
