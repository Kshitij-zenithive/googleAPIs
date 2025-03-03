// internal/domain/models.go
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an authenticated user.
type User struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	GoogleID     string    `gorm:"unique" json:"google_id"` // Google OAuth ID (Unique)
	Email        string    `gorm:"unique" json:"email"`
	Name         string    `json:"name"`
	Picture      string    `json:"picture"`
	AccessToken  string    `json:"-"` // Don't expose in JSON responses
	RefreshToken string    `json:"-"` // Don't expose
	ExpiresAt    time.Time `json:"-"` // Don't expose
}

// Meeting represents a scheduled meeting.
type Meeting struct {
	gorm.Model
	ID              uint      `gorm:"primaryKey" json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	EventID         string    `json:"event_id"`                  // Google Calendar Event ID
	AttendeesString string    `gorm:"column:attendees" json:"-"` // Comma-separated attendees (for storage)
	Attendees       []string  `gorm:"-" json:"attendees"`        // Attendees (for API response)
	CreatedBy       string    `json:"created_by"`                // Email of the user who created the meeting
}

// Attendee represents a participant in a meeting.
type Attendee struct {
	gorm.Model
	MeetingID uint   `json:"meeting_id"`
	Email     string `json:"email"`
}

// AfterFind is a GORM hook that runs after fetching a Meeting.
func (m *Meeting) AfterFind(tx *gorm.DB) (err error) {
	if m.AttendeesString != "" {
		m.Attendees = splitAttendees(m.AttendeesString) // Convert to slice
	}
	return
}

// splitAttendees splits a comma-separated string of attendees into a slice.
func splitAttendees(attendeesStr string) []string {
	return strings.Split(attendeesStr, ",")
}
