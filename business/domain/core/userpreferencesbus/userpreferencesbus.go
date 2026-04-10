package userpreferencesbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("user preference not found")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Upsert(ctx context.Context, pref UserPreference) error
	Delete(ctx context.Context, userID uuid.UUID, key string) error
	QueryByUser(ctx context.Context, userID uuid.UUID) ([]UserPreference, error)
	QueryByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (UserPreference, error)
}

// Business manages the set of APIs for user preference access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a user preferences business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Business value replacing the Storer
// value with a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:    b.log,
		storer: storer,
	}, nil
}

// Set upserts a single preference for a user.
func (b *Business) Set(ctx context.Context, np NewUserPreference) (UserPreference, error) {
	ctx, span := otel.AddSpan(ctx, "business.userpreferencesbus.set")
	defer span.End()

	now := time.Now()

	pref := UserPreference{
		UserID:      np.UserID,
		Key:         np.Key,
		Value:       np.Value,
		UpdatedDate: now,
	}

	if err := b.storer.Upsert(ctx, pref); err != nil {
		return UserPreference{}, fmt.Errorf("set: %w", err)
	}

	return pref, nil
}

// Get retrieves a single preference by user ID and key.
func (b *Business) Get(ctx context.Context, userID uuid.UUID, key string) (UserPreference, error) {
	ctx, span := otel.AddSpan(ctx, "business.userpreferencesbus.get")
	defer span.End()

	pref, err := b.storer.QueryByUserAndKey(ctx, userID, key)
	if err != nil {
		return UserPreference{}, fmt.Errorf("get: %w", err)
	}

	return pref, nil
}

// GetAll retrieves all preferences for a user.
func (b *Business) GetAll(ctx context.Context, userID uuid.UUID) ([]UserPreference, error) {
	ctx, span := otel.AddSpan(ctx, "business.userpreferencesbus.getall")
	defer span.End()

	prefs, err := b.storer.QueryByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getall: %w", err)
	}

	return prefs, nil
}

// Delete removes a single preference by user ID and key.
func (b *Business) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	ctx, span := otel.AddSpan(ctx, "business.userpreferencesbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, userID, key); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
