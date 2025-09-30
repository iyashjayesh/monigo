package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iyashjayesh/monigo"
)

// statusCapturingResponseWriter captures the status code from the underlying handler
type statusCapturingResponseWriter struct {
	gin.ResponseWriter
	statusCode int
}

func (w *statusCapturingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func main() {
	// Initialize MoniGo without starting the dashboard
	monigoInstance := &monigo.Monigo{
		ServiceName:             "gin-integration-example",
		DataPointsSyncFrequency: "5m",
		DataRetentionPeriod:     "7d",
		TimeZone:                "Local",
		CustomBaseAPIPath:       "/monigo/api/v1", // Custom API path
	}

	// Initialize MoniGo (this sets up metrics collection but doesn't start the dashboard)
	monigoInstance.Initialize()

	// Create Gin router
	r := gin.Default()

	// Add your own routes first (these take priority)
	r.GET("/api/users", usersHandler)
	r.POST("/api/orders", ordersHandler)
	r.GET("/health", healthHandler)

	// Get MoniGo unified handler that handles both API and static files
	unifiedHandler := monigo.GetUnifiedHandler("/monigo/api/v1")

	// Register specific routes for MoniGo with custom handlers
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
	log.Println("MoniGo dashboard available at: http://localhost:8080/")
	log.Println("MoniGo API available at: http://localhost:8080/monigo/api/v1/")
	log.Println("Your API available at: http://localhost:8080/api/")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func usersHandler(c *gin.Context) {
	// Trace this function for monitoring
	monigo.TraceFunction(func() {
		// Simulate some work
		_ = make([]byte, 1024*1024) // 1MB allocation
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Users endpoint",
		"count":   42,
	})
}

func ordersHandler(c *gin.Context) {
	// Trace this function for monitoring
	monigo.TraceFunction(func() {
		// Simulate CPU intensive work
		for i := 0; i < 1000000; i++ {
			_ = i * i
		}
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Orders endpoint",
		"count":   15,
	})
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "gin-integration-example",
	})
}
