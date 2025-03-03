package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "github.com/gorilla/mux"
	"google-calendar-api/internal/config"
)

func main() {
	// Load configuration.  This is done *before* calling InitializeApp.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration Error: %v", err)
	}

	// Initialize the application using Wire.  This is the *only* place
	// where we manually create anything. The rest is handled by Wire.
	app, err := InitializeApp(context.Background(), &cfg)
	if err != nil {
		log.Fatalf("❌ Application Initialization Error: %v", err)
	}

	// Create HTTP server.
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      app.Router, // Use the router from the initialized app
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine.
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("🚀 Server is running on port %s", cfg.ServerPort)
		serverErrors <- srv.ListenAndServe()
	}()

	// Graceful shutdown.
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

	// Close the database connection (obtained from the initialized app).
	if err := app.CloseDB(); err != nil {
		log.Printf("⚠️ Error closing database: %v", err)
	}

	log.Println("✅ Server shutdown completed")
}
