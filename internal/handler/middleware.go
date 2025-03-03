// internal/handler/middleware.go
package handler

import (
	"context"
	"google-calendar-api/internal/service"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a type-safe key for storing user information in the request context.
type contextKey string

const userKey contextKey = "user"

// AuthMiddleware validates authentication tokens from request headers or cookies.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Check "Authorization" header for a Bearer token
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		// If no token found in header, check cookies
		if tokenString == "" {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Unauthorized: No valid authentication token", http.StatusUnauthorized)
				return
			}
			tokenString = cookie.Value
		}

		// Validate token and set user info in request context
		ctx, err := h.validateAndSetContext(r.Context(), tokenString) //Pass Context
		if err != nil {
			log.Printf("Authentication failed: %v", err) // Log the error
			http.Error(w, "Unauthorized: Invalid authentication token", http.StatusUnauthorized)
			return
		}

		// Proceed to the next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateAndSetContext verifies the given token and returns a new request context containing user details.
func (h *Handler) validateAndSetContext(ctx context.Context, tokenString string) (context.Context, error) {

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return h.config.JWTSecret, nil
	})

	if err != nil || !token.Valid {
		return ctx, err
	}

	email, ok := (*claims)["email"].(string) // Directly check for the "email" claim
	if !ok {
		return ctx, jwt.ErrTokenInvalidClaims
	}

	// Create and return a context with the UserInfo.  This is now MUCH simpler.
	userInfo := service.UserInfo{ //Using service.UserInfo struct.
		Email: email,
	}

	newCtx := context.WithValue(ctx, userKey, userInfo)

	return newCtx, nil
}

// loggingMiddleware logs incoming HTTP requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📌 %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware catches panics and responds with an internal server error.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("💥 Panic recovered: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
