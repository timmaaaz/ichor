package basicauthapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/domain/http/basicauthapi"
	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/usercache"
	"github.com/timmaaaz/ichor/business/domain/core/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/foundation/keystore"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

func Test_BasicAuth(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_BasicAuth")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, login(db, sd), "login")
	unitest.Run(t, loginInvalid(db, sd), "login_invalid")
	unitest.Run(t, refresh(db, sd), "refresh")
	unitest.Run(t, logout(db, sd), "logout")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 2, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	// Store the password for use in tests
	// Note: We'll need to update the user's password directly in the database
	// since UpdateUser might not have a PasswordHash field
	testPassword := "testpass123"

	_, err = busDomain.User.Update(ctx, usrs[0], userbus.UpdateUser{
		Password: &testPassword,
	})
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("updating user password: %w", err)
	}

	// Convert userbus.User to unitest.User if they're different types
	// If unitest.SeedData expects different type, adjust accordingly
	sd := unitest.SeedData{
		Users: make([]unitest.User, len(usrs)),
	}

	// Convert users - adjust based on your actual types
	for i, u := range usrs {
		sd.Users[i] = unitest.User{
			User: u, // Assuming unitest.User wraps userbus.User
		}
	}

	return sd, nil
}

// =============================================================================

func createTestApp(db *dbtest.Database) (*web.App, *auth.Auth) {
	// Create logger
	log := logger.NewWithEvents(os.Stdout, logger.LevelInfo, "TEST",
		func(ctx context.Context) string { return "" },
		logger.Events{})

	// Create keystore
	ks := keystore.New()

	// Then load keys into it (returns number loaded, not the keystore)
	numKeys, err := ks.LoadByFileSystem(os.DirFS("../../../../zarf/keys"))
	if err != nil {
		panic(fmt.Sprintf("loading keystore: %v", err))
	}
	if numKeys == 0 {
		panic("no keys loaded from zarf/keys")
	}

	// Create auth with proper Config struct
	authCfg := auth.Config{
		Log:       log,
		DB:        db.DB,
		KeyLookup: ks,
		Issuer:    "test-issuer",
	}

	authSvc, err := auth.New(authCfg)
	if err != nil {
		panic(fmt.Sprintf("creating auth: %v", err))
	}

	// Create web app
	logger := func(ctx context.Context, msg string, args ...any) {
		log.Info(ctx, msg, args...)
	}

	app := web.NewApp(
		logger,
		nil, // tracer not needed for tests
		mid.Logger(log),
		mid.Errors(log),
		mid.Metrics(),
		mid.Panics(),
	)

	// Configure basic auth routes
	const tokenKey = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
	const tokenExpiration = 30 * time.Minute

	delegate := delegate.New(log)
	userBus := userbus.NewBusiness(log, delegate, nil, usercache.NewStore(log, userdb.NewStore(log, db.DB), time.Minute))

	cfg := basicauthapi.Config{
		Log:             log,
		Auth:            authSvc,
		DB:              db.DB,
		TokenKey:        tokenKey,
		TokenExpiration: tokenExpiration,
		UserBus:         userBus,
	}
	basicauthapi.Routes(app, cfg)

	return app, authSvc
}

func createTestAuth(db *dbtest.Database) (*auth.Auth, *keystore.KeyStore, string) {
	log := logger.NewWithEvents(os.Stdout, logger.LevelInfo, "TEST",
		func(ctx context.Context) string { return "" },
		logger.Events{})

	ks := keystore.New()
	numKeys, err := ks.LoadByFileSystem(os.DirFS("../../../../zarf/keys"))
	if err != nil {
		panic(fmt.Sprintf("loading keystore: %v", err))
	}
	if numKeys == 0 {
		panic("no keys loaded from zarf/keys")
	}

	authCfg := auth.Config{
		Log:       log,
		DB:        db.DB,
		KeyLookup: ks,
		Issuer:    "test-issuer",
	}

	authSvc, err := auth.New(authCfg)
	if err != nil {
		panic(fmt.Sprintf("creating auth: %v", err))
	}

	const tokenKey = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
	return authSvc, ks, tokenKey
}

// =============================================================================

func login(db *dbtest.Database, sd unitest.SeedData) []unitest.Table {
	app, _ := createTestApp(db)

	table := []unitest.Table{
		{
			Name:    "valid_credentials",
			ExpResp: http.StatusOK,
			ExcFunc: func(ctx context.Context) any {
				// Access the underlying User from unitest.User
				userEmail := sd.Users[0].Email.Address
				if sd.Users[0].User.Email.Address != "" {
					userEmail = sd.Users[0].User.Email.Address
				}

				loginReq := basicauthapi.LoginRequest{
					Email:    userEmail,
					Password: "testpass123",
				}

				body, err := json.Marshal(loginReq)
				if err != nil {
					return err
				}

				req := httptest.NewRequest(http.MethodPost, "/api/auth/basic/login", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					return w.Code
				}

				var resp basicauthapi.LoginResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					return err
				}

				// Validate response has required fields
				if resp.Token == "" {
					return "missing token"
				}

				userID := sd.Users[0].ID.String()
				if sd.Users[0].User.ID.String() != "" {
					userID = sd.Users[0].User.ID.String()
				}

				if resp.UserID != userID {
					return "user ID mismatch"
				}
				if resp.Email != userEmail {
					return "email mismatch"
				}

				return http.StatusOK
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func loginInvalid(db *dbtest.Database, sd unitest.SeedData) []unitest.Table {
	app, _ := createTestApp(db)

	table := []unitest.Table{
		{
			Name:    "invalid_password",
			ExpResp: http.StatusUnauthorized,
			ExcFunc: func(ctx context.Context) any {
				userEmail := sd.Users[0].Email.Address
				if sd.Users[0].User.Email.Address != "" {
					userEmail = sd.Users[0].User.Email.Address
				}

				loginReq := basicauthapi.LoginRequest{
					Email:    userEmail,
					Password: "wrongpassword",
				}

				body, err := json.Marshal(loginReq)
				if err != nil {
					return err
				}

				req := httptest.NewRequest(http.MethodPost, "/api/auth/basic/login", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				return w.Code
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "invalid_email",
			ExpResp: http.StatusUnauthorized,
			ExcFunc: func(ctx context.Context) any {
				loginReq := basicauthapi.LoginRequest{
					Email:    "nonexistent@example.com",
					Password: "testpass123",
				}

				body, err := json.Marshal(loginReq)
				if err != nil {
					return err
				}

				req := httptest.NewRequest(http.MethodPost, "/api/auth/basic/login", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				return w.Code
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "disabled_user",
			ExpResp: http.StatusUnauthorized,
			ExcFunc: func(ctx context.Context) any {
				// Disable the second user
				enabled := false
				_, err := db.BusDomain.User.Update(ctx, sd.Users[1].User, userbus.UpdateUser{
					Enabled: &enabled,
				})
				if err != nil {
					return err
				}

				userEmail := sd.Users[1].User.Email.Address
				loginReq := basicauthapi.LoginRequest{
					Email:    userEmail,
					Password: "testpass123",
				}

				body, err := json.Marshal(loginReq)
				if err != nil {
					return err
				}

				req := httptest.NewRequest(http.MethodPost, "/api/auth/basic/login", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				return w.Code
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func refresh(db *dbtest.Database, sd unitest.SeedData) []unitest.Table {
	app, authSvc := createTestApp(db)

	const tokenKey = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	table := []unitest.Table{
		{
			Name:    "refresh_valid_token",
			ExpResp: http.StatusOK,
			ExcFunc: func(ctx context.Context) any {
				userID := sd.Users[0].ID.String()
				if sd.Users[0].User.ID.String() != "" {
					userID = sd.Users[0].User.ID.String()
				}

				// Generate a token that's close to expiry
				claims := auth.Claims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject:   userID,
						Issuer:    authSvc.Issuer(),
						ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(25 * time.Minute)),
						IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
					},
					Roles: []string{"ADMIN"},
				}

				token, err := authSvc.GenerateToken(tokenKey, claims)
				if err != nil {
					return err
				}

				refreshReq := basicauthapi.RefreshRequest{
					Token: token,
				}

				body, err := json.Marshal(refreshReq)
				if err != nil {
					return err
				}

				req := httptest.NewRequest(http.MethodPost, "/api/auth/basic/refresh", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					return w.Code
				}

				var resp basicauthapi.LoginResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					return err
				}

				// Validate we got a new token
				if resp.Token == "" || resp.Token == token {
					return "token not refreshed"
				}

				return http.StatusOK
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "refresh_not_eligible",
			ExpResp: http.StatusBadRequest,
			ExcFunc: func(ctx context.Context) any {
				userID := sd.Users[0].ID.String()
				if sd.Users[0].User.ID.String() != "" {
					userID = sd.Users[0].User.ID.String()
				}

				// Generate a token that's not close to expiry
				claims := auth.Claims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject:   userID,
						Issuer:    authSvc.Issuer(),
						ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(2 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
					},
					Roles: []string{"USER"}, // Use "USER" or whatever role constant you have
				}

				token, err := authSvc.GenerateToken(tokenKey, claims)
				if err != nil {
					return err
				}

				refreshReq := basicauthapi.RefreshRequest{
					Token: token,
				}

				body, err := json.Marshal(refreshReq)
				if err != nil {
					return err
				}

				req := httptest.NewRequest(http.MethodPost, "/api/auth/basic/refresh", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				return w.Code
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func logout(db *dbtest.Database, sd unitest.SeedData) []unitest.Table {
	app, authSvc := createTestApp(db)
	tokenKey := "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	table := []unitest.Table{
		{
			Name: "logout_valid_token",
			ExpResp: struct {
				Status  int
				Message string
			}{
				Status:  http.StatusOK,
				Message: "logged out",
			},
			ExcFunc: func(ctx context.Context) any {
				userID := sd.Users[0].ID.String()
				if sd.Users[0].User.ID.String() != "" {
					userID = sd.Users[0].User.ID.String()
				}

				// Generate a valid token
				claims := auth.Claims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject:   userID,
						Issuer:    authSvc.Issuer(),
						ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
						IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
					},
					Roles: []string{"USER"},
				}

				token, err := authSvc.GenerateToken(tokenKey, claims)
				if err != nil {
					return err
				}

				req := httptest.NewRequest(http.MethodPost, "/api/auth/basic/logout", nil)
				req.Header.Set("Authorization", "Bearer "+token)

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					return struct {
						Status  int
						Message string
					}{
						Status:  w.Code,
						Message: "",
					}
				}

				var resp struct {
					Message string `json:"message"`
				}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					return err
				}

				return struct {
					Status  int
					Message string
				}{
					Status:  http.StatusOK,
					Message: resp.Message,
				}
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "logout_invalid_token",
			ExpResp: http.StatusUnauthorized,
			ExcFunc: func(ctx context.Context) any {
				req := httptest.NewRequest(http.MethodPost, "/api/auth/basic/logout", nil)
				req.Header.Set("Authorization", "Bearer invalid-token")

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				return w.Code
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
