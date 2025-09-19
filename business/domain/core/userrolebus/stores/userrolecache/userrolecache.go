package userrolecache

import (
	"context"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of apis for roles cache access.
type Store struct {
	log    *logger.Logger
	storer userrolebus.Storer
	cache  *sturdyc.Client[userrolebus.UserRole]
}

// NewStore constructs the api for data and caching access.
func NewStore(log *logger.Logger, storer userrolebus.Storer, ttl time.Duration) *Store {
	const capacity = 10000
	const numShards = 10
	const evictionPercentage = 10

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[userrolebus.UserRole](capacity, numShards, ttl, evictionPercentage),
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (userrolebus.Storer, error) {
	return s.storer.NewWithTx(tx)
}

// Create inserts a new role into the database.
func (s *Store) Create(ctx context.Context, r userrolebus.UserRole) error {
	if err := s.storer.Create(ctx, r); err != nil {
		return err
	}

	s.writeCache(r)

	return nil
}

// Update replaces a role document in the database.
func (s *Store) Update(ctx context.Context, r userrolebus.UserRole) error {
	if err := s.storer.Update(ctx, r); err != nil {
		return err
	}

	s.writeCache(r)

	return nil
}

// Delete removes a role from the database.
func (s *Store) Delete(ctx context.Context, role userrolebus.UserRole) error {
	if err := s.storer.Delete(ctx, role); err != nil {
		return err
	}
	s.deleteCache(role)

	return nil
}

// Query retrieves a list of roles from the database.
func (s *Store) Query(ctx context.Context, filter userrolebus.QueryFilter, orderBy order.By, page page.Page) ([]userrolebus.UserRole, error) {
	return s.storer.Query(ctx, filter, orderBy, page)
}

// Count retrieves the total number of roles from the database.
func (s *Store) Count(ctx context.Context, filter userrolebus.QueryFilter) (int, error) {
	return s.storer.Count(ctx, filter)
}

// QueryByID retrieves a role from the database.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (userrolebus.UserRole, error) {
	bus, exists := s.readCache(id.String())
	if exists {
		return bus, nil
	}

	bus, err := s.storer.QueryByID(ctx, id)
	if err != nil {
		return userrolebus.UserRole{}, err
	}

	s.writeCache(bus)

	return bus, nil
}

// QueryByUserID retrieves a role from the database.
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]userrolebus.UserRole, error) {
	ur, err := s.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return []userrolebus.UserRole{}, err
	}
	for _, r := range ur {
		s.writeCache(r)
	}

	return ur, nil
}

// readCache performs a safe search in the cache for the specified key.
func (s *Store) readCache(key string) (userrolebus.UserRole, bool) {
	usr, exists := s.cache.Get(key)
	if !exists {
		return userrolebus.UserRole{}, false
	}

	return usr, true
}

// writeCache performs a safe write to the cache for the specified userrolebus.
func (s *Store) writeCache(bus userrolebus.UserRole) {
	s.cache.Set(bus.ID.String(), bus)
}

// deleteCache performs a safe removal from the cache for the specified userrolebus.
func (s *Store) deleteCache(bus userrolebus.UserRole) {
	s.cache.Delete(bus.ID.String())
}
