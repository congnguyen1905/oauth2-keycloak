package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"auth-service/config"
	"auth-service/pkg/models"
	"auth-service/pkg/storage"
	authjwt "auth-service/pkg/jwt"
)

// AuthService handles authentication business logic
type AuthService struct {
	config      *config.Config
	storage     storage.Storage
	jwtValidator *authjwt.Validator
	httpClient  *resty.Client
	sessionTTL  time.Duration
}

// NewAuthService creates a new AuthService instance
func NewAuthService(cfg *config.Config, store storage.Storage) *AuthService {
	return &AuthService{
		config:      cfg,
		storage:     store,
		sessionTTL:  5 * time.Minute,
		httpClient:  resty.New(),
		jwtValidator: authjwt.NewValidator(
			cfg.Keycloak.URL,
			cfg.Keycloak.Realm,
			cfg.Keycloak.JWKSURL,
			cfg.JWT.VerifySignature,
			cfg.JWT.Audience,
		),
	}
}

// Login authenticates user with Keycloak and creates a session
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.UserSession, error) {
	// Call Keycloak token endpoint
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", 
		s.config.Keycloak.URL, s.config.Keycloak.Realm)
	
	resp, err := s.httpClient.R().
		SetFormData(map[string]string{
			"grant_type":    "password",
			"client_id":     s.config.Keycloak.ClientID,
			"client_secret": s.config.Keycloak.ClientSecret,
			"username":      req.Username,
			"password":      req.Password,
		}).
		Post(tokenURL)
	
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf("authentication failed")
	}
	
	// Parse token response
	var tokenResponse map[string]interface{}
	if err := resp.Result(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response")
	}
	
	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid access token")
	}
	
	// Decode JWT to extract claims WITHOUT calling Keycloak for verification
	// This is fast and doesn't require network call to Keycloak
	claims, err := s.jwtValidator.DecodeUnverified(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}
	
	// Extract user info from claims
	userID, username, email, roles := authjwt.ExtractUserInfo(claims)
	
	// Create session
	session := &models.UserSession{
		SessionID:   generateSessionID(),
		UserID:      userID,
		Username:    username,
		Email:       email,
		Roles:       roles,
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(s.sessionTTL),
		CreatedAt:   time.Now(),
		Claims:      claims,
	}
	
	// Save session to storage
	if err := s.storage.Save(ctx, session, s.sessionTTL); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}
	
	return session, nil
}

// ValidateSession validates a session ID and returns the session
func (s *AuthService) ValidateSession(ctx context.Context, sessionID string) (*models.ValidateSessionResponse, error) {
	session, err := s.storage.Get(ctx, sessionID)
	if err != nil {
		return &models.ValidateSessionResponse{
			Valid: false,
			Error: "Failed to retrieve session",
		}, nil
	}
	
	if session == nil {
		return &models.ValidateSessionResponse{
			Valid: false,
			Error: "Session not found or expired",
		}, nil
	}
	
	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.storage.Delete(ctx, sessionID)
		return &models.ValidateSessionResponse{
			Valid: false,
			Error: "Session expired",
		}, nil
	}
	
	// Extend session TTL
	s.storage.UpdateTTL(ctx, sessionID, s.sessionTTL)
	session.ExpiresAt = time.Now().Add(s.sessionTTL)
	
	return &models.ValidateSessionResponse{
		Valid:   true,
		Session: session,
	}, nil
}

// Logout removes a session
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.storage.Delete(ctx, sessionID)
}

// Introspect checks if a session is active (for backend services)
func (s *AuthService) Introspect(ctx context.Context, sessionID string) (*models.IntrospectResponse, error) {
	session, err := s.storage.Get(ctx, sessionID)
	if err != nil || session == nil {
		return &models.IntrospectResponse{Active: false}, nil
	}
	
	if time.Now().After(session.ExpiresAt) {
		s.storage.Delete(ctx, sessionID)
		return &models.IntrospectResponse{Active: false}, nil
	}
	
	// Extend session
	s.storage.UpdateTTL(ctx, sessionID, s.sessionTTL)
	
	return &models.IntrospectResponse{
		Active:    true,
		UserID:    session.UserID,
		Username:  session.Username,
		Email:     session.Email,
		Roles:     session.Roles,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// GetSession retrieves a session by ID
func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*models.UserSession, error) {
	return s.storage.Get(ctx, sessionID)
}

// SetSessionTTL sets the session TTL
func (s *AuthService) SetSessionTTL(ttl time.Duration) {
	s.sessionTTL = ttl
}
