package userpreferencesapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
)

// App manages the set of app layer APIs for user preferences.
type App struct {
	userPreferencesBus *userpreferencesbus.Business
}

// NewApp constructs a user preferences app API for use.
func NewApp(userPreferencesBus *userpreferencesbus.Business) *App {
	return &App{
		userPreferencesBus: userPreferencesBus,
	}
}

// Set upserts a single preference for a user.
func (a *App) Set(ctx context.Context, userID uuid.UUID, key string, app NewUserPreference) (UserPreference, error) {
	np := userpreferencesbus.NewUserPreference{
		UserID: userID,
		Key:    key,
		Value:  app.Value,
	}

	pref, err := a.userPreferencesBus.Set(ctx, np)
	if err != nil {
		return UserPreference{}, fmt.Errorf("set: %w", err)
	}

	return ToAppUserPreference(pref), nil
}

// Get retrieves a single preference by user ID and key.
func (a *App) Get(ctx context.Context, userID uuid.UUID, key string) (UserPreference, error) {
	pref, err := a.userPreferencesBus.Get(ctx, userID, key)
	if err != nil {
		if errors.Is(err, userpreferencesbus.ErrNotFound) {
			return UserPreference{}, errs.New(errs.NotFound, err)
		}
		return UserPreference{}, fmt.Errorf("get: %w", err)
	}

	return ToAppUserPreference(pref), nil
}

// GetAll retrieves all preferences for a user.
func (a *App) GetAll(ctx context.Context, userID uuid.UUID) (UserPreferences, error) {
	prefs, err := a.userPreferencesBus.GetAll(ctx, userID)
	if err != nil {
		return UserPreferences{}, fmt.Errorf("getall: %w", err)
	}

	return UserPreferences{
		Items: ToAppUserPreferences(prefs),
	}, nil
}

// Delete removes a single preference by user ID and key.
func (a *App) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	if err := a.userPreferencesBus.Delete(ctx, userID, key); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
