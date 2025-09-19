package rolecache

import (
	"context"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of apis for roles cache access.
type Store struct {
	log    *logger.Logger
	storer rolebus.Storer
	cache  *sturdyc.Client[rolebus.Role]
}

// NewStore constructs the api for data and caching access.
func NewStore(log *logger.Logger, storer rolebus.Storer, ttl time.Duration) *Store {
	const capacity = 10000
	const numShards = 10
	const evictionPercentage = 10

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[rolebus.Role](capacity, numShards, ttl, evictionPercentage),
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (rolebus.Storer, error) {
	return s.storer.NewWithTx(tx)
}

// Create inserts a new role into the database.
func (s *Store) Create(ctx context.Context, r rolebus.Role) error {
	if err := s.storer.Create(ctx, r); err != nil {
		return err
	}

	s.writeCache(r)

	return nil
}

// Update replaces a role document in the database.
func (s *Store) Update(ctx context.Context, r rolebus.Role) error {
	if err := s.storer.Update(ctx, r); err != nil {
		return err
	}

	s.writeCache(r)

	return nil
}

// Delete removes a role from the database.
func (s *Store) Delete(ctx context.Context, role rolebus.Role) error {
	if err := s.storer.Delete(ctx, role); err != nil {
		return err
	}
	s.deleteCache(role)

	return nil
}

// Query retrieves a list of roles from the database.
func (s *Store) Query(ctx context.Context, filter rolebus.QueryFilter, orderBy order.By, page page.Page) ([]rolebus.Role, error) {
	return s.storer.Query(ctx, filter, orderBy, page)
}

// Count retrieves the total number of roles from the database.
func (s *Store) Count(ctx context.Context, filter rolebus.QueryFilter) (int, error) {
	return s.storer.Count(ctx, filter)
}

// QueryByID retrieves a role from the database.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (rolebus.Role, error) {
	bus, exists := s.readCache(id.String())
	if exists {
		return bus, nil
	}

	bus, err := s.storer.QueryByID(ctx, id)
	if err != nil {
		return rolebus.Role{}, err
	}

	s.writeCache(bus)

	return bus, nil
}

// QueryAll retrieves all roles from the database.
func (s *Store) QueryAll(ctx context.Context) ([]rolebus.Role, error) {
	return s.storer.QueryAll(ctx)
}

// QueryByIDs retrieves a list of roles from the database, using cache where possible.
func (s *Store) QueryByIDs(ctx context.Context, ids []uuid.UUID) ([]rolebus.Role, error) {
	if len(ids) == 0 {
		return []rolebus.Role{}, nil
	}

	// First, try to get roles from cache
	var foundRoles []rolebus.Role
	var missingIDs []uuid.UUID

	for _, id := range ids {
		idStr := id.String()
		if role, exists := s.readCache(idStr); exists {
			foundRoles = append(foundRoles, role)
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	// If all roles were in cache, return them
	if len(missingIDs) == 0 {
		return foundRoles, nil
	}

	// Fetch missing roles from database
	dbRoles, err := s.storer.QueryByIDs(ctx, missingIDs)
	if err != nil {
		return nil, err
	}

	// Add newly fetched roles to cache
	for _, role := range dbRoles {
		s.writeCache(role)
	}

	// Combine cached and newly fetched roles
	return append(foundRoles, dbRoles...), nil
}

// readCache performs a safe search in the cache for the specified key.
func (s *Store) readCache(key string) (rolebus.Role, bool) {
	usr, exists := s.cache.Get(key)
	if !exists {
		return rolebus.Role{}, false
	}

	return usr, true
}

// writeCache performs a safe write to the cache for the specified rolebus.
func (s *Store) writeCache(bus rolebus.Role) {
	s.cache.Set(bus.ID.String(), bus)
}

// deleteCache performs a safe removal from the cache for the specified rolebus.
func (s *Store) deleteCache(bus rolebus.Role) {
	s.cache.Delete(bus.ID.String())
}
