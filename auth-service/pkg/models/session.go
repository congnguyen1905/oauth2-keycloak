package models

import "time"

// UserSession represents a user session stored after authentication
type UserSession struct {
	SessionID   string                 `json:"session_id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Roles       []string               `json:"roles"`
	AccessToken string                 `json:"access_token"`
	ExpiresAt   time.Time              `json:"expires_at"`
	CreatedAt   time.Time              `json:"created_at"`
	Claims      map[string]interface{} `json:"claims,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ValidateSessionRequest represents a session validation request
type ValidateSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// ValidateSessionResponse represents a session validation response
type ValidateSessionResponse struct {
	Valid   bool         `json:"valid"`
	Session *UserSession `json:"session,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// IntrospectResponse represents an introspection response
type IntrospectResponse struct {
	Active    bool      `json:"active"`
	UserID    string    `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Email     string    `json:"email,omitempty"`
	Roles     []string  `json:"roles,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}
