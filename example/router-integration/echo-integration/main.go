package main

import (
	"log"
	"net/http"

	"github.com/iyashjayesh/monigo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize MoniGo without starting the dashboard
	monigoInstance := &monigo.Monigo{
		ServiceName:             "echo-integration-example",
		DataPointsSyncFrequency: "5m",
		DataRetentionPeriod:     "7d",
		TimeZone:                "Local",
		CustomBaseAPIPath:       "/monigo/api/v1", // Custom API path
	}

	// Initialize MoniGo (this sets up metrics collection but doesn't start the dashboard)
	monigoInstance.Initialize()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Add your own routes first (these take priority)
	e.GET("/api/users", usersHandler)
	e.POST("/api/orders", ordersHandler)
	e.GET("/health", healthHandler)

	// Get MoniGo unified handler that handles both API and static files
	unifiedHandler := monigo.GetUnifiedHandler("/monigo/api/v1")
	e.Any("/*", echo.WrapHandler(http.HandlerFunc(unifiedHandler)))

	log.Println("Server starting on :8080")
	log.Println("MoniGo dashboard available at: http://localhost:8080/")
	log.Println("MoniGo API available at: http://localhost:8080/monigo/api/v1/")
	log.Println("Your API available at: http://localhost:8080/api/")

	if err := e.Start(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func usersHandler(c echo.Context) error {
	// Trace this function for monitoring
	monigo.TraceFunction(func() {
		// Simulate some work
		_ = make([]byte, 1024*1024) // 1MB allocation
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Users endpoint",
		"count":   42,
	})
}

func ordersHandler(c echo.Context) error {
	// Trace this function for monitoring
	monigo.TraceFunction(func() {
		// Simulate CPU intensive work
		for i := 0; i < 1000000; i++ {
			_ = i * i
		}
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Orders endpoint",
		"count":   15,
	})
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "echo-integration-example",
	})
}
