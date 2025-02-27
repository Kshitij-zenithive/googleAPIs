package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"google-calendar-api/internal/service"
)

// CreateEventRequest represents the request body for creating an event.
type CreateEventRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	Attendees   []string `json:"attendees"` // Use a slice of strings
}

// CreateEvent handles the creation of a new Google Calendar event.
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	//Get User Info
	userInfo, ok := r.Context().Value(userKey).(service.UserInfo)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[ERROR] Failed to decode request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	// Validate start and end times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}
	if startTime.After(endTime) || startTime.Equal(endTime) {
		http.Error(w, "start time must be before end time", http.StatusBadRequest)
		return
	}

	// Validate attendees (basic email format check)
	for _, email := range req.Attendees {
		if !isValidEmail(email) { // Implement a proper email validation function
			http.Error(w, fmt.Sprintf("Invalid attendee email: %s", email), http.StatusBadRequest)
			return
		}
	}

	// Create event using the service layer
	eventID, err := h.eventService.CreateEvent(r.Context(), service.CreateEventInput{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Attendees:   req.Attendees,
		CreatedBy:   userInfo.Email, // Use email from the validated token
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create event: %v", err) // More detailed logging
		//Handle specific errors (like token refresh failures) if possible
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Event created successfully", "event_id": eventID})
}

// ListEvents fetches upcoming meetings from both Google Calendar and the database.
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {

	//Get User Info
	userInfo, ok := r.Context().Value(userKey).(service.UserInfo)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	events, err := h.eventService.ListEvents(r.Context(), userInfo.Email) //Pass Email
	if err != nil {
		log.Printf("[ERROR] Failed to list events: %v", err) // Log the actual error
		http.Error(w, "Failed to list events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"events": events})
}

// Helper Functions

// isValidEmail performs a basic email format check.
func isValidEmail(email string) bool {
	// Basic check for @ and .
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return false
	}

	//Check to prevent very long emails.
	if len(email) > 254 { //RFC 5321
		return false
	}
	return true
}
