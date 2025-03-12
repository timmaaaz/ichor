package permissionscache

import (
	"context"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of apis for permissions cache access.
type Store struct {
	log    *logger.Logger
	storer permissionsbus.Storer
	cache  *sturdyc.Client[permissionsbus.UserPermissions]
}

// NewStore constructs the api for data and caching access.
func NewStore(log *logger.Logger, storer permissionsbus.Storer, ttl time.Duration) *Store {
	const capacity = 10000
	const numShards = 10
	const evictionPercentage = 10

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[permissionsbus.UserPermissions](capacity, numShards, ttl, evictionPercentage),
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (permissionsbus.Storer, error) {
	return s.storer.NewWithTx(tx)
}

// QueryUsrPermissions retrieves the permissions for the specified user.
func (s *Store) QueryUserPermissions(ctx context.Context, userID uuid.UUID) (permissionsbus.UserPermissions, error) {
	cachedPerms, ok := s.readCache(userID.String())
	if ok {
		return cachedPerms, nil
	}

	perms, err := s.storer.QueryUserPermissions(ctx, userID)
	if err != nil {
		return permissionsbus.UserPermissions{}, err
	}

	s.writeCache(perms)

	return perms, nil
}

// readCache performs a safe search in the cache for the specified key.
func (s *Store) readCache(key string) (permissionsbus.UserPermissions, bool) {
	perms, exists := s.cache.Get(key)
	if !exists {
		return permissionsbus.UserPermissions{}, false
	}

	return perms, true
}

// writeCache performs a safe write to the cache for the specified userbus.
func (s *Store) writeCache(bus permissionsbus.UserPermissions) {
	s.cache.Set(bus.UserID.String(), bus)
}

// deleteCache performs a safe removal from the cache for the specified userbus.
func (s *Store) deleteCache(bus userbus.User) {
	s.cache.Delete(bus.ID.String())
	s.cache.Delete(bus.Email.Address)
}
