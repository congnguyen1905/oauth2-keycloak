package service

import (
	"github.com/google/uuid"
)

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return uuid.New().String()
}
