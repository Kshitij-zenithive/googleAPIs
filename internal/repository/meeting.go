// internal/repository/meeting.go
package repository

import (
	"context"
	"google-calendar-api/internal/domain"
	"strings"
	"time"

	"gorm.io/gorm"
)

type meetingRepo struct {
	db *gorm.DB
}

// NewMeetingRepository creates a new MeetingRepository instance.
func NewMeetingRepository(db *gorm.DB) MeetingRepository {
	return &meetingRepo{db}
}

func (r *meetingRepo) CreateMeeting(ctx context.Context, meeting *domain.Meeting) error {
	// Convert attendees slice to a comma-separated string
	attendeesString := strings.Join(meeting.Attendees, ",")
	meeting.AttendeesString = attendeesString // Store in the database field

	tx := r.db.WithContext(ctx).Begin() // Start a transaction
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(meeting).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error // Commit the transaction
}

func (r *meetingRepo) ListMeetingsByUser(ctx context.Context, userEmail string, startTime, endTime time.Time) ([]domain.Meeting, error) {
	var meetings []domain.Meeting
	err := r.db.WithContext(ctx).
		Where("created_by = ? AND start_time >= ? AND end_time <= ?", userEmail, startTime, endTime).
		Find(&meetings).Error
	return meetings, err
}

func (r *meetingRepo) GetMeetingByID(ctx context.Context, id uint) (*domain.Meeting, error) {
	var meeting domain.Meeting
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&meeting).Error
	if err != nil {
		return nil, err // Handle not found appropriately
	}
	return &meeting, nil
}
