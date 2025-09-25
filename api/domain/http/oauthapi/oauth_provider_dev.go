package oauthapi

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/markbates/goth"
	"golang.org/x/oauth2"
)

// DevProvider implements goth.Provider for development environments
type DevProvider struct {
	name        string
	callbackURL string
	debug       bool
}

// DevSession implements goth.Session for development
type DevSession struct {
	AuthURL     string
	AccessToken string
	Email       string
	UserID      string
	Name        string
	Role        string
	ExpiresAt   time.Time
}

// NewDevelopmentProvider creates a new development provider
func NewDevelopmentProvider(callbackURL string) *DevProvider {
	return &DevProvider{
		name:        "development",
		callbackURL: callbackURL,
	}
}

// Provider interface methods

func (d *DevProvider) Name() string {
	return d.name
}

func (d *DevProvider) SetName(name string) {
	d.name = name
}

func (d *DevProvider) BeginAuth(state string) (goth.Session, error) {
	// Construct the full callback URL properly
	callbackPath := fmt.Sprintf("%s/api/auth/development/callback?state=%s", d.callbackURL, state)

	return &DevSession{
		AuthURL:   callbackPath,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (d *DevProvider) UnmarshalSession(data string) (goth.Session, error) {
	sess := &DevSession{}
	err := json.Unmarshal([]byte(data), sess)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (d *DevProvider) FetchUser(session goth.Session) (goth.User, error) {
	sess, ok := session.(*DevSession)
	if !ok {
		return goth.User{}, fmt.Errorf("invalid session type for development provider")
	}

	// If no data set, use defaults
	if sess.Email == "" {
		sess.Email = "dev@example.com"
		sess.UserID = "dev-user-001"
		sess.Name = "Development User"
	}

	user := goth.User{
		UserID:      sess.UserID,
		Provider:    d.name,
		Email:       sess.Email,
		Name:        sess.Name,
		AccessToken: sess.AccessToken,
		ExpiresAt:   sess.ExpiresAt,
		RawData: map[string]interface{}{
			"development": true,
			"role":        sess.Role,
		},
	}

	return user, nil
}

func (d *DevProvider) Debug(debug bool) {
	d.debug = debug
}

func (d *DevProvider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: fmt.Sprintf("dev-token-%d", time.Now().Unix()),
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(24 * time.Hour),
	}, nil
}

func (d *DevProvider) RefreshTokenAvailable() bool {
	return false
}

// Session interface methods

func (s *DevSession) GetAuthURL() (string, error) {
	return s.AuthURL, nil
}

func (s *DevSession) Marshal() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *DevSession) Authorize(provider goth.Provider, params goth.Params) (string, error) {
	// Extract user info from query params
	email := params.Get("email")
	if email == "" {
		email = "dev@example.com"
	}

	name := params.Get("name")
	if name == "" {
		name = "Development User"
	}

	role := params.Get("role")
	if role == "" {
		role = "USER"
	}

	userID := params.Get("user_id")
	if userID == "" {
		userID = fmt.Sprintf("dev-%s-%d", email, time.Now().Unix())
	}

	s.Email = email
	s.Name = name
	s.UserID = userID
	s.Role = role
	s.AccessToken = fmt.Sprintf("dev-token-%d", time.Now().Unix())
	s.ExpiresAt = time.Now().Add(24 * time.Hour)

	return s.AccessToken, nil
}

func (s *DevSession) String() string {
	return s.Marshal()
}
