package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/iyashjayesh/monigo"
	"github.com/labstack/echo/v4"
)

func main() {
	// Rate limit middleware returns (mw, stop) - call stop on shutdown
	rateLimitMW, stop := monigo.RateLimitMiddleware(200, time.Minute)
	defer stop()

	// Initialize MoniGo with security middleware
	monigoInstance := &monigo.Monigo{
		ServiceName:             "echo-secured-example",
		DataPointsSyncFrequency: "5m",
		DataRetentionPeriod:     "7d",
		TimeZone:                "Local",
		CustomBaseAPIPath:       "/monigo/api/v1",

		// Security Configuration
		DashboardMiddleware: []func(http.Handler) http.Handler{
			monigo.BasicAuthMiddleware("admin", "echo-secure-2024"),
			monigo.LoggingMiddleware(),
		},
		APIMiddleware: []func(http.Handler) http.Handler{
			monigo.BasicAuthMiddleware("admin", "echo-secure-2024"),
			rateLimitMW,
		},
	}

	// Initialize MoniGo (this sets up metrics collection but doesn't start the dashboard)
	monigoInstance.Initialize()

	// Create Echo instance
	e := echo.New()

	// Add your own routes first (these take priority)
	e.GET("/api/users", usersHandler)
	e.POST("/api/orders", ordersHandler)
	e.GET("/health", healthHandler)

	// Register MoniGo secured handlers with Echo
	unifiedHandler := monigo.GetSecuredUnifiedHandler(monigoInstance, "/monigo/api/v1")

	// Register API handlers first (more specific)
	e.Any("/monigo/api/v1/*", echo.WrapHandler(http.HandlerFunc(unifiedHandler)))

	// Register dashboard handler (catch-all for everything else)
	e.Any("/*", echo.WrapHandler(http.HandlerFunc(unifiedHandler)))

	log.Println("Server starting on :8080")
	log.Println("MoniGo dashboard available at: http://localhost:8080/")
	log.Println("Username: admin, Password: echo-secure-2024")
	log.Println("IP whitelist removed for testing - accessible from any IP")
	log.Println("MoniGo API available at: http://localhost:8080/monigo/api/v1/")
	log.Println("Your API available at: http://localhost:8080/api/")

	if err := e.Start(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func usersHandler(c echo.Context) error {
	// Trace this function for monitoring
	monigo.TraceFunction(context.Background(), func() {
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
	monigo.TraceFunction(context.Background(), func() {
		// Simulate some work
		_ = make([]byte, 512*1024) // 512KB allocation
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Orders endpoint",
		"count":   15,
	})
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "healthy",
	})
}
