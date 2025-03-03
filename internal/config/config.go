// internal/config/config.go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Config holds application configuration.
type Config struct {
	DatabaseURL        string
	ServerPort         string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTSecret          []byte         // Changed to byte slice
	CSRFSecret         []byte         // Changed to byte slice
	OAuthConfig        *oauth2.Config // OAuth configuration
	Env                string
}

// LoadConfig initializes configuration from environment variables.
func LoadConfig() (Config, error) {
	// Load .env file for non-production environments.
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ Warning: No .env file found")
		}
	}

	//Required Environment variables.
	requiredEnvs := []string{
		"DB_URL",
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
		"GOOGLE_REDIRECT_URL",
		"JWT_SECRET",
		"CSRF_SECRET",
	}
	for _, envVar := range requiredEnvs {
		if os.Getenv(envVar) == "" {
			return Config{}, fmt.Errorf("%s is missing in environment variables", envVar)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	conf := Config{
		DatabaseURL:        os.Getenv("DB_URL"),
		ServerPort:         port,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		JWTSecret:          []byte(os.Getenv("JWT_SECRET")),  // Store as byte slice
		CSRFSecret:         []byte(os.Getenv("CSRF_SECRET")), // Store as byte slice
		Env:                os.Getenv("ENV"),
	}

	conf.OAuthConfig = &oauth2.Config{
		ClientID:     conf.GoogleClientID,
		ClientSecret: conf.GoogleClientSecret,
		RedirectURL:  conf.GoogleRedirectURL,
		Scopes: []string{
			"openid",
			"email",
			"profile",
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/calendar.events",
		},
		Endpoint: google.Endpoint,
	}
	return conf, nil
}
