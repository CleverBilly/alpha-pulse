package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"alpha-pulse/backend/internal/auth"
	"alpha-pulse/backend/internal/handler"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestProtectedAPIRequiresAuthWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	authHandler, authMiddleware := newTestAuth(t)
	r := newTestRouterWithAuth(t, db, authHandler, authMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/api/market-snapshot?symbol=BTCUSDT&interval=1h&limit=24", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status, got=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLoginSessionAndProtectedRouteFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	authHandler, authMiddleware := newTestAuth(t)
	r := newTestRouterWithAuth(t, db, authHandler, authMiddleware)

	loginBody := bytes.NewBufferString(`{"username":"alpha-admin","password":"alpha-pass"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	r.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login success, got=%d body=%s", loginRec.Code, loginRec.Body.String())
	}

	sessionCookie := loginRec.Result().Header.Get("Set-Cookie")
	if sessionCookie == "" {
		t.Fatal("expected login response to set auth cookie")
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	sessionReq.Header.Set("Cookie", sessionCookie)
	sessionRec := httptest.NewRecorder()
	r.ServeHTTP(sessionRec, sessionReq)

	if sessionRec.Code != http.StatusOK {
		t.Fatalf("expected session endpoint success, got=%d body=%s", sessionRec.Code, sessionRec.Body.String())
	}

	var sessionPayload apiEnvelope[map[string]any]
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionPayload); err != nil {
		t.Fatalf("decode session response failed: %v", err)
	}
	if sessionPayload.Data["authenticated"] != true {
		t.Fatalf("expected authenticated session payload, got=%#v", sessionPayload.Data)
	}

	protectedReq := httptest.NewRequest(http.MethodGet, "/api/market-snapshot?symbol=BTCUSDT&interval=1h&limit=24", nil)
	protectedReq.Header.Set("Cookie", sessionCookie)
	protectedRec := httptest.NewRecorder()
	r.ServeHTTP(protectedRec, protectedReq)

	if protectedRec.Code != http.StatusOK {
		t.Fatalf("expected protected route to succeed after login, got=%d body=%s", protectedRec.Code, protectedRec.Body.String())
	}
}

func newTestAuth(t *testing.T) (*handler.AuthHandler, gin.HandlerFunc) {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("alpha-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate bcrypt password hash failed: %v", err)
	}

	authService, err := auth.NewService(auth.Options{
		Enabled:       true,
		Username:      "alpha-admin",
		PasswordHash:  string(passwordHash),
		SessionSecret: "super-secret",
		SessionTTL:    24 * time.Hour,
		CookieName:    "alpha_pulse_session",
		CookieSecure:  false,
	})
	if err != nil {
		t.Fatalf("new auth service failed: %v", err)
	}

	return handler.NewAuthHandler(authService), handler.RequireAuth(authService)
}
