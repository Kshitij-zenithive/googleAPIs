package main

import (
	"context"
	"google-calendar-api/internal/config"
	"google-calendar-api/internal/handler"
	"google-calendar-api/internal/repository"
	"google-calendar-api/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "google-calendar-api/internal/domain" // Import for side effects (migrations)
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration Error: %v", err)
	}

	// Initialize database connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database Connection Error: %v", err)
	}
	sqlDB, err := db.DB() // Get underlying *sql.DB
	if err != nil {
		log.Fatalf("Failed to get *sql.DB: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil { // Correctly close the connection
			log.Printf("⚠️ Error closing database: %v", err)
		}
	}()

	// Run database migrations (using AutoMigrate for simplicity here)
	if err := repository.MigrateDB(db); err != nil {
		log.Fatalf("❌ migration error %v\n", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	meetingRepo := repository.NewMeetingRepository(db)

	// Initialize services
	authService := service.NewAuthService(&cfg, userRepo) // Pass config to AuthService
	eventService := service.NewEventService(meetingRepo, userRepo, &cfg)

	// Initialize handlers
	h := handler.NewHandler(authService, eventService, &cfg) //Pass Config

	// Initialize router
	router := mux.NewRouter()

	// Setup routes and middleware (defined in handler package)
	h.RegisterRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("🚀 Server is running on port %s", cfg.ServerPort)
		serverErrors <- srv.ListenAndServe()
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("❌ Server error: %v", err)
	case sig := <-stop:
		log.Printf("🛑 Received signal %v. Initiating shutdown...", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Error during server shutdown: %v", err)
	}

	log.Println("✅ Server shutdown completed")
}
