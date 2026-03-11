package auth

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestServiceAuthenticatesCredentialsAndVerifiesSessionToken(t *testing.T) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte("alpha-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate bcrypt hash failed: %v", err)
	}

	service, err := NewService(Options{
		Enabled:         true,
		Username:        "alpha-admin",
		PasswordHash:    string(hashBytes),
		SessionSecret:   "super-secret",
		SessionTTL:      2 * time.Hour,
		CookieName:      "alpha_pulse_session",
		CookieDomain:    "",
		CookieSecure:    false,
		AllowedOrigins:  []string{"http://localhost:3000"},
	})
	if err != nil {
		t.Fatalf("new service failed: %v", err)
	}

	if err := service.Authenticate("alpha-admin", "alpha-pass"); err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if err := service.Authenticate("alpha-admin", "wrong-pass"); err == nil {
		t.Fatal("expected invalid password to be rejected")
	}
	if err := service.Authenticate("wrong-user", "alpha-pass"); err == nil {
		t.Fatal("expected invalid username to be rejected")
	}

	token, expiry, err := service.IssueSessionToken("alpha-admin", time.Now())
	if err != nil {
		t.Fatalf("issue session token failed: %v", err)
	}
	if expiry.Before(time.Now().Add(119 * time.Minute)) {
		t.Fatalf("unexpected expiry: %s", expiry)
	}

	session, err := service.VerifySessionToken(token, time.Now())
	if err != nil {
		t.Fatalf("verify session token failed: %v", err)
	}
	if session.Username != "alpha-admin" {
		t.Fatalf("unexpected session username: %s", session.Username)
	}
}

func TestServiceRejectsExpiredAndTamperedTokens(t *testing.T) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte("alpha-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate bcrypt hash failed: %v", err)
	}

	service, err := NewService(Options{
		Enabled:         true,
		Username:        "alpha-admin",
		PasswordHash:    string(hashBytes),
		SessionSecret:   "super-secret",
		SessionTTL:      time.Hour,
		CookieName:      "alpha_pulse_session",
		CookieDomain:    "",
		CookieSecure:    true,
		AllowedOrigins:  []string{"https://app.example.com"},
	})
	if err != nil {
		t.Fatalf("new service failed: %v", err)
	}

	token, _, err := service.IssueSessionToken("alpha-admin", time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatalf("issue session token failed: %v", err)
	}

	if _, err := service.VerifySessionToken(token+"tampered", time.Unix(1_700_000_000, 0)); err == nil {
		t.Fatal("expected tampered token to be rejected")
	}
	if _, err := service.VerifySessionToken(token, time.Unix(1_700_000_000, 0).Add(2*time.Hour)); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestNewServiceRejectsInvalidEnabledConfig(t *testing.T) {
	if _, err := NewService(Options{
		Enabled:      true,
		Username:     "alpha-admin",
		PasswordHash: "",
		SessionSecret:"super-secret",
		SessionTTL:   time.Hour,
		CookieName:   "alpha_pulse_session",
	}); err == nil {
		t.Fatal("expected missing password hash to fail")
	}

	if _, err := NewService(Options{
		Enabled:      true,
		Username:     "alpha-admin",
		PasswordHash: "$2a$10$mockedhashvalue",
		SessionSecret:"",
		SessionTTL:   time.Hour,
		CookieName:   "alpha_pulse_session",
	}); err == nil {
		t.Fatal("expected missing session secret to fail")
	}
}
