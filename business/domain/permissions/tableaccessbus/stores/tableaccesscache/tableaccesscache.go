package tableaccesscache

import (
	"context"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of apis for roles cache access.
type Store struct {
	log    *logger.Logger
	storer tableaccessbus.Storer
	cache  *sturdyc.Client[tableaccessbus.TableAccess]
}

// NewStore constructs the api for data and caching access.
func NewStore(log *logger.Logger, storer tableaccessbus.Storer, ttl time.Duration) *Store {
	const capacity = 10000
	const numShards = 10
	const evictionPercentage = 10

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[tableaccessbus.TableAccess](capacity, numShards, ttl, evictionPercentage),
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (tableaccessbus.Storer, error) {
	return s.storer.NewWithTx(tx)
}

// Create inserts a new role into the database.
func (s *Store) Create(ctx context.Context, r tableaccessbus.TableAccess) error {
	if err := s.storer.Create(ctx, r); err != nil {
		return err
	}

	s.writeCache(r)

	return nil
}

// Update replaces a role document in the database.
func (s *Store) Update(ctx context.Context, r tableaccessbus.TableAccess) error {
	if err := s.storer.Update(ctx, r); err != nil {
		return err
	}

	s.writeCache(r)

	return nil
}

// Delete removes a role from the database.
func (s *Store) Delete(ctx context.Context, role tableaccessbus.TableAccess) error {
	if err := s.storer.Delete(ctx, role); err != nil {
		return err
	}
	s.deleteCache(role)

	return nil
}

// Query retrieves a list of roles from the database.
func (s *Store) Query(ctx context.Context, filter tableaccessbus.QueryFilter, orderBy order.By, page page.Page) ([]tableaccessbus.TableAccess, error) {
	return s.storer.Query(ctx, filter, orderBy, page)
}

// Count retrieves the total number of roles from the database.
func (s *Store) Count(ctx context.Context, filter tableaccessbus.QueryFilter) (int, error) {
	return s.storer.Count(ctx, filter)
}

// QueryByID retrieves a role from the database.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (tableaccessbus.TableAccess, error) {
	bus, exists := s.readCache(id.String())
	if exists {
		return bus, nil
	}

	bus, err := s.storer.QueryByID(ctx, id)
	if err != nil {
		return tableaccessbus.TableAccess{}, err
	}

	s.writeCache(bus)

	return bus, nil
}

// QueryByRoleIDs retrieves table access entries by role IDs with caching.
func (s *Store) QueryByRoleIDs(ctx context.Context, roleIDs []uuid.UUID) ([]tableaccessbus.TableAccess, error) {
	// First query the database to get all matching table access entries
	tableAccesses, err := s.storer.QueryByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	// Cache all retrieved entries
	for _, access := range tableAccesses {
		s.writeCache(access)
	}

	return tableAccesses, nil
}

// readCache performs a safe search in the cache for the specified key.
func (s *Store) readCache(key string) (tableaccessbus.TableAccess, bool) {
	usr, exists := s.cache.Get(key)
	if !exists {
		return tableaccessbus.TableAccess{}, false
	}

	return usr, true
}

// writeCache performs a safe write to the cache for the specified tableaccessbus.
func (s *Store) writeCache(bus tableaccessbus.TableAccess) {
	s.cache.Set(bus.ID.String(), bus)
}

// deleteCache performs a safe removal from the cache for the specified tableaccessbus.
func (s *Store) deleteCache(bus tableaccessbus.TableAccess) {
	s.cache.Delete(bus.ID.String())
}
