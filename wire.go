// wire.go
//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"database/sql"

	"google-calendar-api/internal/config"
	"google-calendar-api/internal/handler"
	"google-calendar-api/internal/repository"
	"google-calendar-api/internal/service"

	"github.com/google/wire"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// App struct to hold the top-level application components.
type App struct {
	Router *mux.Router
	DB     *gorm.DB
	SqlDB  *sql.DB // Add this to close the raw *sql.DB
}

// CloseDB closes the database connection.
func (a *App) CloseDB() error {
	if a.SqlDB != nil {
		return a.SqlDB.Close()
	}
	return nil
}

// InitializeApp is the Wire injector function.
func InitializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	wire.Build(
		NewDB,
		repository.NewUserRepository,
		repository.NewMeetingRepository,
		service.NewAuthService,
		service.NewEventService,
		handler.NewHandler,
		NewRouter,
		NewApp,
	)
	return &App{}, nil // This return is replaced by Wire.
}

// NewApp creates a new App instance.  This is a *provider*.
func NewApp(router *mux.Router, db *gorm.DB, sqlDB *sql.DB) *App {
	return &App{Router: router, DB: db, SqlDB: sqlDB}
}

// NewRouter creates a new mux.Router. This is a *provider*.
func NewRouter(h *handler.Handler) *mux.Router {
	router := mux.NewRouter()
	h.RegisterRoutes(router)
	return router
}

// NewDB creates anew gorm.DB connection. This is a *provider*.
func NewDB(cfg *config.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, nil, err // Return nil for both if there's an error
	}

	//Run Migrations
	if err := repository.MigrateDB(db); err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB() //get the standard sql.DB
	if err != nil {
		return nil, nil, err
	}
	return db, sqlDB, nil
}
