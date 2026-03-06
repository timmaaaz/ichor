// Package oauthapi maintains the web based api for oauth support.
package oauthapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// validProviders is the allowlist of known OAuth provider names. Any provider
// name not in this map is rejected before being passed to Gothic.
var validProviders = map[string]bool{
	"google":      true,
	"development": true,
}

type api struct {
	log             *logger.Logger
	auth            *auth.Auth
	userBus         *userbus.Business
	blocklist       *auth.Blocklist
	store           sessions.Store
	tokenKey        string
	uiAdminRedirect string
	uiLoginRedirect string
	tokenExpiration time.Duration
}

func newAPI(cfg Config) *api {
	// Fix 2: Startup assertion — refuse to start if production config uses dev defaults.
	if cfg.Environment == "production" && cfg.StoreKey == "dev-session-key-32-bytes-long!!!" {
		panic("oauthapi: production environment detected with development StoreKey — refusing to start")
	}

	// Set up providers based on environment. Google is optional in all
	// environments — internal deployments may rely solely on basic auth.
	if cfg.Environment == "production" {
		// Production: dev provider is never registered (security invariant).
		if cfg.GoogleKey != "" && cfg.GoogleSecret != "" {
			goth.UseProviders(
				google.New(cfg.GoogleKey, cfg.GoogleSecret, cfg.Callback),
			)
		}
	} else {
		// Development/Staging - add dev provider
		providers := []goth.Provider{
			NewDevelopmentProvider(cfg.Callback),
		}
		if cfg.GoogleKey != "" && cfg.GoogleSecret != "" {
			providers = append(providers,
				google.New(cfg.GoogleKey, cfg.GoogleSecret, cfg.Callback))
		}
		goth.UseProviders(providers...)
	}

	// Fix 5+6: Session cookie — Secure flag tied to environment, SameSite Lax,
	// MaxAge 15 minutes (sufficient for OAuth handshake, eliminates 30-day window).
	store := sessions.NewCookieStore([]byte(cfg.StoreKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   900, // 15 minutes
		HttpOnly: true,
		Secure:   cfg.Environment == "production",
		SameSite: http.SameSiteLaxMode,
	}

	gothic.Store = store

	// Fix 9: Provider allowlist — reject unknown provider names before Gothic sees them.
	gothic.GetProviderName = func(r *http.Request) (string, error) {
		segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(segments) >= 3 && segments[0] == "api" && segments[1] == "auth" {
			provider := segments[2]
			if provider == "callback" && len(segments) >= 4 {
				provider = segments[3]
			}
			if !validProviders[provider] {
				return "", fmt.Errorf("unknown oauth provider: %q", provider)
			}
			return provider, nil
		}

		return "", errors.New("provider not found in path")
	}

	return &api{
		log:             cfg.Log,
		auth:            cfg.Auth,
		userBus:         cfg.UserBus,
		blocklist:       cfg.Blocklist,
		store:           store,
		tokenKey:        cfg.TokenKey,
		uiAdminRedirect: cfg.UIAdminRedirect,
		uiLoginRedirect: cfg.UILoginRedirect,
		tokenExpiration: cfg.TokenExpiration,
	}
}

func (a *api) authenticate(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func (a *api) authCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		a.log.Error(r.Context(), "completing user auth: %s", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	sess, err := a.store.Get(r, "user-metadata")
	if err != nil {
		a.log.Error(r.Context(), "get session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sess.Values["user"] = user

	if err := sess.Save(r, w); err != nil {
		a.log.Error(r.Context(), "save session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Fix 3: Look up the user in the database to get their actual roles rather
	// than granting Admin unconditionally.
	addr, err := mail.ParseAddress(user.Email)
	if err != nil {
		a.log.Error(r.Context(), "oauth callback: invalid email from provider", "email", user.Email, "err", err)
		http.Error(w, "Unauthorized: invalid email", http.StatusForbidden)
		return
	}

	dbUser, err := a.userBus.QueryByEmail(r.Context(), *addr)
	if err != nil {
		a.log.Error(r.Context(), "oauth callback: user not found", "email", user.Email, "err", err)
		http.Error(w, "Unauthorized: user not registered", http.StatusForbidden)
		return
	}

	if !dbUser.Enabled {
		a.log.Error(r.Context(), "oauth callback: disabled user attempted login", "email", user.Email)
		http.Error(w, "Unauthorized: account disabled", http.StatusForbidden)
		return
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   dbUser.ID.String(),
			Issuer:    a.auth.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(a.tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: userbus.ParseRolesToString(dbUser.Roles),
	}

	token, err := a.auth.GenerateToken(a.tokenKey, claims)
	if err != nil {
		a.log.Error(r.Context(), "generating token: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, a.uiAdminRedirect+token, http.StatusFound)
}

func (a *api) logout(w http.ResponseWriter, r *http.Request) {
	// Revoke the JWT so it cannot be used after logout. Use Authenticate (not
	// ParseClaims) to verify the token signature before blocklisting — this
	// prevents an attacker from blocklisting arbitrary JTIs with forged tokens.
	if a.blocklist != nil {
		if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
			if claims, err := a.auth.Authenticate(r.Context(), authHeader); err == nil && claims.ID != "" {
				a.blocklist.Add(claims.ID, claims.ExpiresAt.Time)
			}
		}
	}

	sess, err := a.store.Get(r, "user-metadata")
	if err != nil {
		a.log.Error(r.Context(), "get session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Logout user by invalidating their session data.
	sess.Values["user"] = nil

	if err := sess.Save(r, w); err != nil {
		a.log.Error(r.Context(), "save session: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := gothic.Logout(w, r); err != nil {
		a.log.Error(r.Context(), "gothic logout: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, a.uiLoginRedirect, http.StatusFound)
}
