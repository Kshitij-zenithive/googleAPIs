package service

import (
	"context"
	"time"

	"golang.org/x/oauth2"
)

// AuthService defines the interface for authentication operations.
type AuthService interface {
	HandleGoogleCallback(ctx context.Context, token *oauth2.Token) (string, error) // Returns JWT
}

// UserInfo represents user information extracted from a token
type UserInfo struct {
	Email string
}

// EventService defines the interface for event-related operations.
type EventService interface {
	CreateEvent(ctx context.Context, input CreateEventInput) (string, error) // Returns event ID
	ListEvents(ctx context.Context, userEmail string) ([]EventOutput, error)
}

// CreateEventInput represents the input for creating an event.
type CreateEventInput struct {
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Attendees   []string // Use a slice of strings
	CreatedBy   string
}

type EventOutput struct {
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Attendees   []string
	EventId     string // Google Calendar event ID.
	CreatedBy   string
}

type AttendeeOutput struct {
	Email string
}
