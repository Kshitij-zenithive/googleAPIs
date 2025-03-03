// internal/service/auth.go
package service

import (
	"context"
	"fmt"
	"google-calendar-api/internal/config"
	"google-calendar-api/internal/domain"
	"google-calendar-api/internal/repository"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type authService struct {
	userRepo     repository.UserRepository
	oauthConfig  *oauth2.Config // Use oauth2.Config
	jwtSecret    []byte
	oidcProvider *oidc.Provider // OIDC Provider
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(cfg *config.Config, userRepo repository.UserRepository) *authService {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		//This should stop the execution of the app.
		log.Fatalf("❌ Failed to create OIDC provider: %v\n", err)
	}
	return &authService{
		userRepo:     userRepo,
		oauthConfig:  cfg.OAuthConfig, // Use directly from config
		jwtSecret:    cfg.JWTSecret,
		oidcProvider: provider,
	}
}

// HandleGoogleCallback handles the OAuth2 callback from Google, creates/updates the user, and generates a JWT.
func (s *authService) HandleGoogleCallback(ctx context.Context, token *oauth2.Token) (string, error) {
	// Extract ID Token from the token response
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", fmt.Errorf("no id_token field in oauth2 token")
	}

	// Verify and decode the ID Token
	verifier := s.oidcProvider.Verifier(&oidc.Config{ClientID: s.oauthConfig.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Decode token claims to get user details
	var claims struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("failed to parse ID token claims: %w", err)
	}

	// Ensure required fields are present
	if claims.Sub == "" || claims.Email == "" {
		return "", fmt.Errorf("userinfo missing required fields: %+v", claims)
	}
	// Check if user exists
	user, err := s.userRepo.GetUserByGoogleID(ctx, claims.Sub)
	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}

	if user == nil {
		// Create new user
		newUser := &domain.User{
			ID:           uuid.New(), // Use UUID
			GoogleID:     claims.Sub,
			Email:        claims.Email,
			Name:         claims.Name,
			Picture:      claims.Picture,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry,
		}
		if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user
		user.AccessToken = token.AccessToken
		user.ExpiresAt = token.Expiry
		if token.RefreshToken != "" {
			user.RefreshToken = token.RefreshToken
		}
		user.Name = claims.Name
		user.Picture = claims.Picture //Update user data.
		if err := s.userRepo.UpdateUser(ctx, user); err != nil {
			return "", fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Generate JWT
	jwtToken, err := s.generateJWT(claims.Email) // Generate JWT with email
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return jwtToken, nil
}

// generateJWT creates a JWT for the user.
func (s *authService) generateJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,                                 // Include user's email
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
