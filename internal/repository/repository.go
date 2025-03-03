// internal/repository/repository.go
package repository

import (
	"context"
	"google-calendar-api/internal/domain"
	"time"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

// MeetingRepository defines the interface for meeting data access.
type MeetingRepository interface {
	CreateMeeting(ctx context.Context, meeting *domain.Meeting) error
	ListMeetingsByUser(ctx context.Context, userEmail string, startTime, endTime time.Time) ([]domain.Meeting, error)
	GetMeetingByID(ctx context.Context, id uint) (*domain.Meeting, error) // Added GetMeetingByID
}

// MigrateDB performs database migrations.
func MigrateDB(db *gorm.DB) error {
	return db.AutoMigrate(&domain.User{}, &domain.Meeting{}, &domain.Attendee{})
}
