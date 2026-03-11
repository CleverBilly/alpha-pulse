package handler

import (
	"net/http"
	"time"

	authsvc "alpha-pulse/backend/internal/auth"
	"alpha-pulse/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *authsvc.Service
	now     func() time.Time
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authSessionResponse struct {
	Enabled       bool   `json:"enabled"`
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username,omitempty"`
}

func NewAuthHandler(service *authsvc.Service) *AuthHandler {
	return &AuthHandler{
		service: service,
		now:     time.Now,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	if h.service == nil || !h.service.Enabled() {
		c.JSON(http.StatusNotFound, utils.Error(http.StatusNotFound, "auth disabled"))
		return
	}

	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(http.StatusBadRequest, "invalid login payload"))
		return
	}

	if err := h.service.Authenticate(request.Username, request.Password); err != nil {
		c.JSON(http.StatusUnauthorized, utils.Error(http.StatusUnauthorized, "invalid username or password"))
		return
	}

	token, expiresAt, err := h.service.IssueSessionToken(request.Username, h.now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(http.StatusInternalServerError, "issue session failed"))
		return
	}

	http.SetCookie(c.Writer, h.service.BuildSessionCookie(token, expiresAt))
	c.JSON(http.StatusOK, utils.Success(authSessionResponse{
		Enabled:       true,
		Authenticated: true,
		Username:      request.Username,
	}))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if h.service == nil || !h.service.Enabled() {
		c.JSON(http.StatusOK, utils.Success(authSessionResponse{
			Enabled:       false,
			Authenticated: false,
		}))
		return
	}

	http.SetCookie(c.Writer, h.service.BuildExpiredCookie())
	c.JSON(http.StatusOK, utils.Success(authSessionResponse{
		Enabled:       true,
		Authenticated: false,
	}))
}

func (h *AuthHandler) Session(c *gin.Context) {
	if h.service == nil || !h.service.Enabled() {
		c.JSON(http.StatusOK, utils.Success(authSessionResponse{
			Enabled:       false,
			Authenticated: true,
		}))
		return
	}

	session, err := h.service.SessionFromRequest(c.Request, h.now())
	if err != nil {
		c.JSON(http.StatusOK, utils.Success(authSessionResponse{
			Enabled:       true,
			Authenticated: false,
		}))
		return
	}

	c.JSON(http.StatusOK, utils.Success(authSessionResponse{
		Enabled:       true,
		Authenticated: true,
		Username:      session.Username,
	}))
}

func RequireAuth(service *authsvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if service == nil || !service.Enabled() {
			c.Next()
			return
		}

		if _, err := service.SessionFromRequest(c.Request, time.Now()); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.Error(http.StatusUnauthorized, "authentication required"))
			return
		}

		c.Next()
	}
}
