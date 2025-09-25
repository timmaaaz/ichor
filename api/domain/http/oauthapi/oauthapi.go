// Package oauthapi maintains the web based api for oauth support.
package oauthapi

import (
	"errors"
	"net/http"
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

type api struct {
	log             *logger.Logger
	auth            *auth.Auth
	store           sessions.Store
	tokenKey        string
	uiAdminRedirect string
	uiLoginRedirect string
	tokenExpiration time.Duration
}

func newAPI(cfg Config) *api {
	// Set up providers based on environment
	if cfg.Environment == "production" {
		goth.UseProviders(
			google.New(cfg.GoogleKey, cfg.GoogleSecret, cfg.Callback),
		)
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

	// Configure Gothic's session store
	store := sessions.NewCookieStore([]byte(cfg.StoreKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	}

	gothic.Store = store

	// Fix the provider extraction
	gothic.GetProviderName = func(r *http.Request) (string, error) {
		// Extract provider from path: /api/auth/{provider}
		segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		// Should be ["api", "auth", "provider"] or ["api", "auth", "provider", "callback"]
		if len(segments) >= 3 && segments[0] == "api" && segments[1] == "auth" {
			provider := segments[2]
			// Remove "callback" if present
			if provider == "callback" && len(segments) >= 4 {
				provider = segments[3]
			}
			return provider, nil
		}

		// Fallback to query parameter
		if provider := r.URL.Query().Get("provider"); provider != "" {
			return provider, nil
		}

		return "", errors.New("provider not found in path")
	}

	return &api{
		log:             cfg.Log,
		auth:            cfg.Auth,
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

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.UserID,
			Issuer:    a.auth.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(a.tokenExpiration)), // Use variable
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{userbus.Roles.Admin.String()},
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
