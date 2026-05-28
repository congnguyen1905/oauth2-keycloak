package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"auth-service/pkg/models"
	"auth-service/pkg/service"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRoutes registers all auth routes
func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/auth/login", h.HandleLogin)
	r.POST("/auth/logout", h.HandleLogout)
	r.POST("/auth/validate", h.HandleValidateSession)
	r.GET("/auth/session/:sessionID", h.HandleGetSession)
	r.POST("/introspect", h.HandleIntrospect)
	r.GET("/health", h.HealthCheck)
}

// HandleLogin handles login requests
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	session, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid username or password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": session.SessionID,
		"user": gin.H{
			"username": session.Username,
			"email":    session.Email,
			"roles":    session.Roles,
		},
	})
}

// HandleLogout handles logout requests
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	var req models.ValidateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.SessionID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "logout_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// HandleValidateSession handles session validation requests
func (h *AuthHandler) HandleValidateSession(c *gin.Context) {
	var req models.ValidateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.ValidateSession(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	if !response.Valid {
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleGetSession handles get session requests
func (h *AuthHandler) HandleGetSession(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Session ID is required",
		})
		return
	}

	session, err := h.authService.GetSession(c.Request.Context(), sessionID)
	if err != nil || session == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Session not found",
		})
		return
	}

	c.JSON(http.StatusOK, session)
}

// HandleIntrospect handles introspection requests (for backend services)
func (h *AuthHandler) HandleIntrospect(c *gin.Context) {
	var req models.ValidateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.Introspect(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "introspection_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck handles health check requests
func (h *AuthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status:    "healthy",
		Timestamp: c.GetString("time"),
		Version:   "1.0.0",
	})
}
