package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAuthDisabled      = errors.New("auth disabled")
	ErrInvalidCredential = errors.New("invalid credential")
	ErrInvalidSession    = errors.New("invalid session")
)

type Options struct {
	Enabled        bool
	Username       string
	PasswordHash   string
	SessionSecret  string
	SessionTTL     time.Duration
	CookieName     string
	CookieDomain   string
	CookieSecure   bool
	AllowedOrigins []string
}

type Session struct {
	Username  string
	ExpiresAt time.Time
}

type Service struct {
	enabled        bool
	username       string
	passwordHash   string
	sessionSecret  []byte
	sessionTTL     time.Duration
	cookieName     string
	cookieDomain   string
	cookieSecure   bool
	allowedOrigins []string
}

func NewService(options Options) (*Service, error) {
	service := &Service{
		enabled:        options.Enabled,
		username:       strings.TrimSpace(options.Username),
		passwordHash:   strings.TrimSpace(options.PasswordHash),
		sessionSecret:  []byte(options.SessionSecret),
		sessionTTL:     options.SessionTTL,
		cookieName:     strings.TrimSpace(options.CookieName),
		cookieDomain:   strings.TrimSpace(options.CookieDomain),
		cookieSecure:   options.CookieSecure,
		allowedOrigins: append([]string(nil), options.AllowedOrigins...),
	}

	if !service.enabled {
		return service, nil
	}
	if service.username == "" {
		return nil, errors.New("auth username is required")
	}
	if service.passwordHash == "" {
		return nil, errors.New("auth password hash is required")
	}
	if len(service.sessionSecret) == 0 {
		return nil, errors.New("auth session secret is required")
	}
	if service.sessionTTL <= 0 {
		return nil, errors.New("auth session ttl must be positive")
	}
	if service.cookieName == "" {
		return nil, errors.New("auth cookie name is required")
	}

	return service, nil
}

func (s *Service) Enabled() bool {
	return s != nil && s.enabled
}

func (s *Service) Authenticate(username, password string) error {
	if !s.Enabled() {
		return ErrAuthDisabled
	}
	if strings.TrimSpace(username) != s.username {
		return ErrInvalidCredential
	}
	if err := bcrypt.CompareHashAndPassword([]byte(s.passwordHash), []byte(password)); err != nil {
		return ErrInvalidCredential
	}
	return nil
}

func (s *Service) IssueSessionToken(username string, now time.Time) (string, time.Time, error) {
	if !s.Enabled() {
		return "", time.Time{}, ErrAuthDisabled
	}

	expiresAt := now.Add(s.sessionTTL)
	payload := fmt.Sprintf("%s:%d", strings.TrimSpace(username), expiresAt.Unix())
	payloadEncoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	signatureEncoded := base64.RawURLEncoding.EncodeToString(s.sign([]byte(payload)))
	return payloadEncoded + "." + signatureEncoded, expiresAt, nil
}

func (s *Service) VerifySessionToken(token string, now time.Time) (Session, error) {
	if !s.Enabled() {
		return Session{}, ErrAuthDisabled
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Session{}, ErrInvalidSession
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Session{}, ErrInvalidSession
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Session{}, ErrInvalidSession
	}
	if !hmac.Equal(signature, s.sign(payload)) {
		return Session{}, ErrInvalidSession
	}

	rawParts := strings.Split(string(payload), ":")
	if len(rawParts) != 2 {
		return Session{}, ErrInvalidSession
	}

	expiresAt, err := time.ParseInLocation(time.RFC3339, "", time.UTC)
	if err == nil {
		_ = expiresAt
	}

	unixExpiry, err := parseUnix(rawParts[1])
	if err != nil {
		return Session{}, ErrInvalidSession
	}
	expiry := time.Unix(unixExpiry, 0)
	if now.After(expiry) {
		return Session{}, ErrInvalidSession
	}

	return Session{
		Username:  rawParts[0],
		ExpiresAt: expiry,
	}, nil
}

func (s *Service) CookieName() string {
	return s.cookieName
}

func (s *Service) CookieDomain() string {
	return s.cookieDomain
}

func (s *Service) CookieSecure() bool {
	return s.cookieSecure
}

func (s *Service) AllowedOrigins() []string {
	return append([]string(nil), s.allowedOrigins...)
}

func (s *Service) SessionFromRequest(request *http.Request, now time.Time) (Session, error) {
	if !s.Enabled() {
		return Session{Username: s.username}, nil
	}

	cookie, err := request.Cookie(s.cookieName)
	if err != nil {
		return Session{}, ErrInvalidSession
	}
	return s.VerifySessionToken(cookie.Value, now)
}

func (s *Service) BuildSessionCookie(token string, expiresAt time.Time) *http.Cookie {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	return &http.Cookie{
		Name:     s.cookieName,
		Value:    token,
		Path:     "/",
		Domain:   s.cookieDomain,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
		Expires:  expiresAt,
	}
}

func (s *Service) BuildExpiredCookie() *http.Cookie {
	return &http.Cookie{
		Name:     s.cookieName,
		Value:    "",
		Path:     "/",
		Domain:   s.cookieDomain,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}

func (s *Service) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, s.sessionSecret)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

func parseUnix(value string) (int64, error) {
	var unix int64
	_, err := fmt.Sscanf(value, "%d", &unix)
	if err != nil {
		return 0, err
	}
	return unix, nil
}
