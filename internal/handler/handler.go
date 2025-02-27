package handler

import (
	"google-calendar-api/internal/config"
	"google-calendar-api/internal/service"
	"net/http"

	"github.com/gorilla/mux"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	authService  service.AuthService
	eventService service.EventService
	config       *config.Config // Add Config
}

// NewHandler creates a new Handler instance.
func NewHandler(authService service.AuthService, eventService service.EventService, cfg *config.Config) *Handler {
	return &Handler{
		authService:  authService,
		eventService: eventService,
		config:       cfg, // Store Config
	}
}

// RegisterRoutes sets up the routes and middleware for the application.
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Global middlewares
	router.Use(loggingMiddleware)
	router.Use(recoveryMiddleware)

	// Public Routes
	router.HandleFunc("/login", h.LoginPage).Methods("GET")
	router.HandleFunc("/auth/google/login", h.GoogleLogin).Methods("GET")
	router.HandleFunc("/auth/google/callback", h.GoogleCallback).Methods("GET")

	// Protected API Routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(h.AuthMiddleware) // Authentication middleware for protected routes
	api.HandleFunc("/dashboard", h.Dashboard).Methods("GET")
	api.HandleFunc("/events", h.CreateEvent).Methods("POST") // /api/events
	api.HandleFunc("/events", h.ListEvents).Methods("GET")   // /api/events

	// Logout Route
	router.HandleFunc("/logout", h.Logout).Methods("GET")

	//static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./internal/web"))))

}
