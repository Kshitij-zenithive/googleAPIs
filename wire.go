//go:build wireinject
// +build wireinject

package main

import (
	"google-calendar-api/internal/config"
	"google-calendar-api/internal/handler"
	"google-calendar-api/internal/repository"
	"google-calendar-api/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// InitializeApp creates all necessary components using Wire.
func InitializeApp(db *gorm.DB, cfg *config.Config) (*handler.Handler, error) {
	wire.Build(
		repository.NewUserRepository,
		repository.NewMeetingRepository,
		service.NewAuthService,  // Added!  Tell wire how to create AuthService
		service.NewEventService, // Added!  Tell wire how to create EventService
		handler.NewHandler,
	)
	return &handler.Handler{}, nil // This is a placeholder; Wire replaces it.
}
