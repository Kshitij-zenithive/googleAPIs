// internal/handler/auth.go
package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"google-calendar-api/internal/service"

	"github.com/google/uuid"
)

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("internal/web/templates/login.html") // Relative Path
	if err != nil {
		log.Printf("Failed to load template: %v", err)
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Message string
	}{
		Message: "Please log in to access your dashboard.",
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated (context should be populated by middleware)
	_, ok := r.Context().Value(userKey).(service.UserInfo)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("internal/web/templates/dashboard.html") // Relative path.
	if err != nil {
		log.Printf("Failed to load template: %v", err)
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// GoogleLogin generates and stores a state, then redirects to Google's OAuth2 endpoint.
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String() // Generate a cryptographically secure random state

	// Store the state in a cookie (for CSRF protection).
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		HttpOnly: true,
		Secure:   h.config.Env == "production", // Use true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Path:     "/", // Important: Set the path to match the callback
	})
	url := h.config.OAuthConfig.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusSeeOther)
}

// GoogleCallback handles the OAuth2 callback from Google.
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stateCookie, err := r.Cookie("oauthstate")
	if err != nil {
		log.Println("❌ State cookie not found:", err)
		http.Error(w, "State cookie not found", http.StatusBadRequest)
		return
	}

	receivedState := r.URL.Query().Get("state")
	if receivedState == "" || receivedState != stateCookie.Value {
		log.Println("❌ Invalid state parameter")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear the state cookie after validation
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.config.Env == "production",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("❌ No code found in request")
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	//Exchange token
	token, err := h.config.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Printf("❌ error exchanging token: %v", err)
		http.Error(w, fmt.Sprintf("Token Exchange Failed: %v", err), http.StatusInternalServerError)
		return
	}
	// Call the AuthService to handle user creation/update and JWT generation
	jwtToken, err := h.authService.HandleGoogleCallback(ctx, token)
	if err != nil {
		log.Printf("❌ Error handling Google callback: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError) // Generic error
		return
	}

	// Store JWT in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.config.Env == "production",
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/api/dashboard", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.config.Env == "production",
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
