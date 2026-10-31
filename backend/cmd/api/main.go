package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/config"
	"backend/database"
	"backend/internal/container"
	_package "backend/package"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/session"
)

///////////////////////////////////////////////////////////////////

func main() {

	log.SetPrefix("[server] ")
	log.SetFlags(log.Ldate | log.Ltime)

	// Loading Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Starting server")
	log.Printf("Environment: %s", cfg.Server.Environment)

	// Connect DB
	db, err := database.Connect(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	log.Println("Database connected successfully")

	// Run Migration
	if err := database.RunMigration(db); err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}
	log.Println("Database migration completed")

	// Initialize Dependency Container
	c := container.NewContainer(db, cfg)
	defer c.Close()
	log.Println("Dependency container initialized")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "The ZINs Beauty and Spa",
		ServerHeader: "Fiber",
		ErrorHandler: _package.FiberErrorHandler, // error ของ Fiber (413, panic, ...) ตอบเป็น ErrorResponse
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.AllowOrigins,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowCredentials: true,
	}))

	app.Use(session.New(session.Config{
		Storage:        nil,
		CookieHTTPOnly: true,
		CookieSecure:   cfg.IsProduction(),
		CookieSameSite: "Lax",

		IdleTimeout:     60 * time.Minute,
		AbsoluteTimeout: 24 * time.Hour,

		Extractor: extractors.FromCookie("the_zins_session"),
	}))

	// Routes
	setupRoute(app, c)

	// Graceful shutdown
	go func() {
		if err := app.Listen(cfg.Server.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server is running on port %s", cfg.Server.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	log.Println("Server stopped gracefully")
}

///////////////////////////////////////////////////////////////////

func setupRoute(app *fiber.App, c *container.Container) {

	app.Get("/", func(ctx fiber.Ctx) error {

		return ctx.JSON(fiber.Map{
			"message": "The ZINs Beauty and Spa API",
			"status":  "Running",
			"version": "1.0.0",
		})

	})

	app.Get("/health", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API Routes
	api := app.Group("/api")

	// Register Authentication Routes
	c.AuthHandler.RegisterRoutes(api)

	// Register Customer Routes
	c.CustomerHandler.RegisterRoutes(api)

	// Register Order Routes
	c.OrderHandler.RegisterRoutes(api)

	// Register Stock Routes
	c.StockHandler.RegisterRoutes(api)

	// Register Stock Log Routes
	c.StockLogHandler.RegisterRoutes(api)

	// Register Course
	c.CourseHandler.RegisterRoutes(api)

	// 404 Handler
	app.Use(func(c fiber.Ctx) error {
		return _package.SendJSON(c, fiber.StatusNotFound,
			_package.ErrorResponse("Route not found", _package.CodeNotFound, c.Method()+" "+c.Path()))
	})

}
