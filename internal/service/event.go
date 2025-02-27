package service

import (
	"context"
	"fmt"
	"google-calendar-api/internal/config"
	"google-calendar-api/internal/domain"
	"google-calendar-api/internal/repository"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type eventService struct {
	meetingRepo repository.MeetingRepository
	userRepo    repository.UserRepository
	oauthConfig *oauth2.Config
}

// NewEventService creates a new EventService instance.
func NewEventService(meetingRepo repository.MeetingRepository, userRepo repository.UserRepository, cfg *config.Config) *eventService {
	return &eventService{
		meetingRepo: meetingRepo,
		userRepo:    userRepo,
		oauthConfig: cfg.OAuthConfig,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, input CreateEventInput) (string, error) {
	//Retrieve User
	user, err := s.userRepo.GetUserByEmail(ctx, input.CreatedBy) // Find user to get credentials
	if err != nil || user == nil {
		log.Printf("❌ User not found or error: %v\n", err)
		return "", fmt.Errorf("user not found")
	}

	//Create an oauth2 token
	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       user.ExpiresAt,
	}

	// Create a Google Calendar service client
	client := s.oauthConfig.Client(ctx, token) // Use context here
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("failed to create calendar service: %w", err)
	}

	// Convert attendee emails into Google Calendar Attendee objects
	var eventAttendees []*calendar.EventAttendee
	for _, email := range input.Attendees {
		eventAttendees = append(eventAttendees, &calendar.EventAttendee{Email: email})
	}

	// Create the event object for Google Calendar
	event := &calendar.Event{
		Summary:     input.Title,
		Description: input.Description,
		Start: &calendar.EventDateTime{
			DateTime: input.StartTime.Format(time.RFC3339),
			TimeZone: input.StartTime.Location().String(), // Use the start time's location
		},
		End: &calendar.EventDateTime{
			DateTime: input.EndTime.Format(time.RFC3339),
			TimeZone: input.EndTime.Location().String(), // Use the end time's location
		},
		Attendees: eventAttendees, // Add attendees
	}

	// Insert event into Google Calendar
	createdEvent, err := service.Events.Insert("primary", event).Do()
	if err != nil {
		log.Printf("❌ Error creating event in Google Calendar %v\n", err)
		// Check if the error is due to token expiry
		if isTokenExpiredError(err) {
			// Try to refresh the token
			log.Println("🔄 Attempting to refresh token...")
			newToken, refreshErr := s.refreshToken(ctx, user)
			if refreshErr != nil {
				return "", fmt.Errorf("failed to refresh token: %w", refreshErr)
			}

			// Update the user with the new token
			user.AccessToken = newToken.AccessToken
			user.RefreshToken = newToken.RefreshToken
			user.ExpiresAt = newToken.Expiry
			if updateErr := s.userRepo.UpdateUser(ctx, user); updateErr != nil {
				return "", fmt.Errorf("failed to update user with new token: %w", updateErr)
			}

			// Retry creating the event with the new token
			client = s.oauthConfig.Client(ctx, newToken)
			service, err := calendar.NewService(ctx, option.WithHTTPClient(client)) // Use context here

			if err != nil {
				return "", fmt.Errorf("failed to create calendar service after refresh: %w", err)
			}
			createdEvent, err = service.Events.Insert("primary", event).Do()
			if err != nil {
				return "", fmt.Errorf("failed to create event after token refresh: %w", err)
			}

		} else {
			return "", fmt.Errorf("failed to create event: %w", err)
		}
	}

	// Store event in the database
	meeting := &domain.Meeting{
		Title:       input.Title,
		Description: input.Description,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		EventID:     createdEvent.Id, // Store Google Calendar event ID
		Attendees:   input.Attendees,
		CreatedBy:   input.CreatedBy,
	}

	if err := s.meetingRepo.CreateMeeting(ctx, meeting); err != nil {
		return "", fmt.Errorf("failed to store event in database: %w", err)
	}

	return createdEvent.Id, nil
}

func (s *eventService) ListEvents(ctx context.Context, userEmail string) ([]EventOutput, error) {
	// Retrieve User by Email.
	user, err := s.userRepo.GetUserByEmail(ctx, userEmail)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Create an oauth2 token.
	token := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       user.ExpiresAt,
	}

	// Create Google Calendar service client.
	client := s.oauthConfig.Client(ctx, token)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client)) // Use context here
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	// Fetch upcoming meetings from Google Calendar for the next 7 days.
	now := time.Now().Format(time.RFC3339)
	weekLater := time.Now().AddDate(0, 0, 7).Format(time.RFC3339)

	events, err := service.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(now).
		TimeMax(weekLater).
		OrderBy("startTime").
		Do()

	if err != nil {
		log.Printf("❌ Error fetching from Google Calendar: %v", err)
		// Check if the error is due to token expiry
		if isTokenExpiredError(err) {
			// Try to refresh the token
			log.Println("🔄 Attempting to refresh token...")
			newToken, refreshErr := s.refreshToken(ctx, user)
			if refreshErr != nil {
				return nil, fmt.Errorf("failed to refresh token: %w", refreshErr)
			}

			// Update the user with the new token
			user.AccessToken = newToken.AccessToken
			user.RefreshToken = newToken.RefreshToken
			user.ExpiresAt = newToken.Expiry

			if updateErr := s.userRepo.UpdateUser(ctx, user); updateErr != nil {
				return nil, fmt.Errorf("failed to update user with refreshed token, %w", updateErr)
			}

			// Retry creating the event with the new token
			client = s.oauthConfig.Client(ctx, newToken)
			service, err = calendar.NewService(ctx, option.WithHTTPClient(client)) // Use context here

			if err != nil {
				return nil, fmt.Errorf("failed to create calendar service after refresh: %w", err)
			}

			events, err = service.Events.List("primary").
				ShowDeleted(false).
				SingleEvents(true).
				TimeMin(now).
				TimeMax(weekLater).
				OrderBy("startTime").
				Do()
			if err != nil {
				return nil, fmt.Errorf("failed to list events after token refresh: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to fetch events from Google Calendar: %w", err)
		}
	}

	//Convert to EventOutput
	var eventOutputs []EventOutput
	for _, item := range events.Items {
		startTime, _ := time.Parse(time.RFC3339, item.Start.DateTime)
		endTime, _ := time.Parse(time.RFC3339, item.End.DateTime)

		var attendees []string
		if item.Attendees != nil {
			for _, a := range item.Attendees {
				attendees = append(attendees, a.Email)
			}
		}

		eventOutputs = append(eventOutputs, EventOutput{
			Title:       item.Summary,
			Description: item.Description,
			StartTime:   startTime,
			EndTime:     endTime,
			Attendees:   attendees,
			EventId:     item.Id,
			CreatedBy:   userEmail, // The user who is listing the events
		})
	}

	return eventOutputs, nil
}

// refreshToken refreshes the access token using the refresh token.
func (s *eventService) refreshToken(ctx context.Context, user *domain.User) (*oauth2.Token, error) {
	tokenSource := s.oauthConfig.TokenSource(ctx, &oauth2.Token{RefreshToken: user.RefreshToken})
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

// isTokenExpiredError checks if the error is due to an expired or invalid token.
func isTokenExpiredError(err error) bool {
	// This is a simplified check.  In a real application, you'd need to inspect
	// the error more thoroughly, possibly checking for specific error codes
	// returned by the Google API.
	return err != nil && (
	// Google uses specific error structures; check for those
	// This is a placeholder; you *must* adapt this to the *actual* error
	// structure returned by the Google API.
	// Example:  strings.Contains(err.Error(), "Token expired") ||
	//           strings.Contains(err.Error(), "Invalid Credentials")
	//
	//  You might need to use `googleapi.CheckResponse` to properly inspect the error:
	//  https://pkg.go.dev/google.golang.org/api/googleapi#CheckResponse
	strings.Contains(err.Error(), "invalid_grant") || // Common OAuth2 error
		strings.Contains(err.Error(), "expired") || //Possible expired
		strings.Contains(err.Error(), "Invalid Credentials")) //Possible invalid
}
