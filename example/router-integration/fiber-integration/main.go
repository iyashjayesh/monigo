package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/iyashjayesh/monigo"
)

func main() {
	// Initialize MoniGo without starting the dashboard
	monigoInstance := &monigo.Monigo{
		ServiceName:             "fiber-integration-example",
		DataPointsSyncFrequency: "5m",
		DataRetentionPeriod:     "7d",
		TimeZone:                "Local",
		CustomBaseAPIPath:       "/monigo/api/v1", // Custom API path
	}

	// Initialize MoniGo (this sets up metrics collection but doesn't start the dashboard)
	monigoInstance.Initialize()

	// Create Fiber app
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Add your own routes first (these take priority)
	app.Get("/api/users", usersHandler)
	app.Post("/api/orders", ordersHandler)
	app.Get("/health", healthHandler)

	// Get MoniGo Fiber handler that handles both API and static files
	fiberHandler := monigo.GetFiberHandler("/monigo/api/v1")
	app.All("*", fiberHandler)

	log.Println("Server starting on :8080")
	log.Println("MoniGo dashboard available at: http://localhost:8080/")
	log.Println("MoniGo API available at: http://localhost:8080/monigo/api/v1/")
	log.Println("Your API available at: http://localhost:8080/api/")

	if err := app.Listen(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func usersHandler(c *fiber.Ctx) error {
	// Trace this function for monitoring
	monigo.TraceFunction(func() {
		// Simulate some work
		_ = make([]byte, 1024*1024) // 1MB allocation
	})

	return c.JSON(fiber.Map{
		"message": "Users endpoint",
		"count":   42,
	})
}

func ordersHandler(c *fiber.Ctx) error {
	// Trace this function for monitoring
	monigo.TraceFunction(func() {
		// Simulate CPU intensive work
		for i := 0; i < 1000000; i++ {
			_ = i * i
		}
	})

	return c.JSON(fiber.Map{
		"message": "Orders endpoint",
		"count":   15,
	})
}

func healthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "fiber-integration-example",
	})
}
