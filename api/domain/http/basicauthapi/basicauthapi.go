package basicauthapi

// api/domain/http/basicauth/basicauth.go

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the request payload for login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (app *LoginRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// LoginResponse represents the response payload for login
type LoginResponse struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (app LoginResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// RefreshRequest represents the request payload for token refresh
type RefreshRequest struct {
	Token string `json:"token" validate:"required"`
}

func (app *RefreshRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// api maintains the set of dependencies for the basic auth endpoints
type api struct {
	log             *logger.Logger
	auth            *auth.Auth
	db              *sqlx.DB
	tokenKey        string
	tokenExpiration time.Duration
	userBus         *userbus.Business
}

type loggedInOutResponse struct {
	Message string `json:"message"`
}

func (app loggedInOutResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// NewAPI constructs a basic auth api for the provided dependencies.
func NewAPI(cfg Config) *api {
	return &api{
		log:             cfg.Log,
		auth:            cfg.Auth,
		db:              cfg.DB,
		tokenKey:        cfg.TokenKey,
		tokenExpiration: cfg.TokenExpiration,
		userBus:         cfg.UserBus,
	}
}

// login handles user authentication with email and password
func (a *api) login(ctx context.Context, r *http.Request) web.Encoder {
	var req LoginRequest
	if err := web.Decode(r, &req); err != nil {
		a.log.Error(ctx, "decoding login request", "error", err)
		return errs.New(errs.InvalidArgument, err)
	}

	email, err := mail.ParseAddress(req.Email)
	if err != nil {
		a.log.Error(ctx, "parsing email address", "error", err)
		return errs.Newf(errs.InvalidArgument, "invalid email format")
	}

	user, err := a.userBus.QueryByEmail(ctx, *email)
	if err != nil {
		a.log.Info(ctx, "login attempt failed", "email", req.Email, "error", err)
		return errs.Newf(errs.Unauthenticated, "invalid credentials")
	}

	if !user.Enabled {
		return errs.Newf(errs.Unauthenticated, "account disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		a.log.Info(ctx, "invalid password", "email", req.Email)
		return errs.Newf(errs.Unauthenticated, "invalid credentials")
	}

	roleStrings := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleStrings[i] = role.String()
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    a.auth.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(a.tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: roleStrings,
	}

	token, err := a.auth.GenerateToken(a.tokenKey, claims)
	if err != nil {
		a.log.Error(ctx, "generating token", "error", err)
		return errs.New(errs.Internal, err)
	}

	a.log.Info(ctx, "user logged in", "user_id", user.ID, "email", user.Email)

	return LoginResponse{
		Token:     token,
		UserID:    user.ID.String(),
		Email:     user.Email.Address,
		Roles:     roleStrings,
		ExpiresAt: time.Now().UTC().Add(a.tokenExpiration),
	}
}

// refresh handles token refresh requests
func (a *api) refresh(ctx context.Context, r *http.Request) web.Encoder {
	var req RefreshRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	claims, err := a.auth.Authenticate(ctx, "Bearer "+req.Token)
	if err != nil {
		return errs.Newf(errs.Unauthenticated, "invalid token")
	}

	timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
	if timeUntilExpiry > 30*time.Minute {
		return errs.Newf(errs.FailedPrecondition, "token not eligible for refresh yet")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid user ID")
	}

	user, err := a.userBus.QueryByID(ctx, userID)
	if err != nil {
		return errs.Newf(errs.NotFound, "user not found")
	}

	if !user.Enabled {
		return errs.Newf(errs.PermissionDenied, "account disabled")
	}

	roleStrings := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleStrings[i] = role.String()
	}

	newClaims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    a.auth.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(a.tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: roleStrings,
	}

	token, err := a.auth.GenerateToken(a.tokenKey, newClaims)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	a.log.Info(ctx, "token refreshed", "user_id", user.ID)

	return LoginResponse{
		Token:     token,
		UserID:    user.ID.String(),
		Email:     user.Email.Address,
		Roles:     roleStrings,
		ExpiresAt: time.Now().UTC().Add(a.tokenExpiration),
	}
}

// logout handles user logout (optional - mainly for audit logging)
func (a *api) logout(ctx context.Context, r *http.Request) web.Encoder {
	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims, err := a.auth.Authenticate(ctx, "Bearer "+token)
	if err != nil {
		return errs.Newf(errs.Unauthenticated, "invalid token")
	}

	a.log.Info(ctx, "user logged out", "user_id", claims.Subject)

	// In a stateful session system, you would invalidate the token here
	// For stateless JWT, the client just discards the token

	return loggedInOutResponse{
		Message: "logged out",
	}
}

// ============================================================================
// Additional helper for password hashing when creating users
// ============================================================================

// HashPassword generates a bcrypt hash for the given password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ============================================================================
// Optional: Session management for stateful sessions
// ============================================================================

// SessionStore interface for session management (optional)
type SessionStore interface {
	Create(ctx context.Context, userID string, token string, expiry time.Duration) error
	Validate(ctx context.Context, token string) (string, error)
	Delete(ctx context.Context, token string) error
}

// InMemorySessionStore is a simple in-memory session store (for development)
type InMemorySessionStore struct {
	sessions map[string]sessionData
}

type sessionData struct {
	UserID    string
	ExpiresAt time.Time
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]sessionData),
	}
}

func (s *InMemorySessionStore) Create(ctx context.Context, userID string, token string, expiry time.Duration) error {
	s.sessions[token] = sessionData{
		UserID:    userID,
		ExpiresAt: time.Now().Add(expiry),
	}
	return nil
}

func (s *InMemorySessionStore) Validate(ctx context.Context, token string) (string, error) {
	session, exists := s.sessions[token]
	if !exists {
		return "", errors.New("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, token)
		return "", errors.New("session expired")
	}
	return session.UserID, nil
}

func (s *InMemorySessionStore) Delete(ctx context.Context, token string) error {
	delete(s.sessions, token)
	return nil
}
