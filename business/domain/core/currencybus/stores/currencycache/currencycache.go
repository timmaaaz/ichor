package currencycache

import (
	"context"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of apis for currencies cache access.
type Store struct {
	log    *logger.Logger
	storer currencybus.Storer
	cache  *sturdyc.Client[currencybus.Currency]
}

// NewStore constructs the api for data and caching access.
func NewStore(log *logger.Logger, storer currencybus.Storer, ttl time.Duration) *Store {
	const capacity = 10000
	const numShards = 10
	const evictionPercentage = 10

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[currencybus.Currency](capacity, numShards, ttl, evictionPercentage),
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (currencybus.Storer, error) {
	return s.storer.NewWithTx(tx)
}

// Create inserts a new currency into the database.
func (s *Store) Create(ctx context.Context, c currencybus.Currency) error {
	if err := s.storer.Create(ctx, c); err != nil {
		return err
	}

	s.writeCache(c)

	return nil
}

// Update replaces a currency document in the database.
func (s *Store) Update(ctx context.Context, c currencybus.Currency) error {
	if err := s.storer.Update(ctx, c); err != nil {
		return err
	}

	s.writeCache(c)

	return nil
}

// Delete removes a currency from the database.
func (s *Store) Delete(ctx context.Context, currency currencybus.Currency) error {
	if err := s.storer.Delete(ctx, currency); err != nil {
		return err
	}
	s.deleteCache(currency)

	return nil
}

// Query retrieves a list of currencies from the database.
func (s *Store) Query(ctx context.Context, filter currencybus.QueryFilter, orderBy order.By, page page.Page) ([]currencybus.Currency, error) {
	return s.storer.Query(ctx, filter, orderBy, page)
}

// Count retrieves the total number of currencies from the database.
func (s *Store) Count(ctx context.Context, filter currencybus.QueryFilter) (int, error) {
	return s.storer.Count(ctx, filter)
}

// QueryByID retrieves a currency from the database.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (currencybus.Currency, error) {
	bus, exists := s.readCache(id.String())
	if exists {
		return bus, nil
	}

	bus, err := s.storer.QueryByID(ctx, id)
	if err != nil {
		return currencybus.Currency{}, err
	}

	s.writeCache(bus)

	return bus, nil
}

// QueryAll retrieves all currencies from the database.
func (s *Store) QueryAll(ctx context.Context) ([]currencybus.Currency, error) {
	return s.storer.QueryAll(ctx)
}

// QueryByIDs retrieves a list of currencies from the database, using cache where possible.
func (s *Store) QueryByIDs(ctx context.Context, ids []uuid.UUID) ([]currencybus.Currency, error) {
	if len(ids) == 0 {
		return []currencybus.Currency{}, nil
	}

	// First, try to get currencies from cache
	var foundCurrencies []currencybus.Currency
	var missingIDs []uuid.UUID

	for _, id := range ids {
		idStr := id.String()
		if currency, exists := s.readCache(idStr); exists {
			foundCurrencies = append(foundCurrencies, currency)
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	// If all currencies were in cache, return them
	if len(missingIDs) == 0 {
		return foundCurrencies, nil
	}

	// Fetch missing currencies from database
	dbCurrencies, err := s.storer.QueryByIDs(ctx, missingIDs)
	if err != nil {
		return nil, err
	}

	// Add newly fetched currencies to cache
	for _, currency := range dbCurrencies {
		s.writeCache(currency)
	}

	// Combine cached and newly fetched currencies
	return append(foundCurrencies, dbCurrencies...), nil
}

// readCache performs a safe search in the cache for the specified key.
func (s *Store) readCache(key string) (currencybus.Currency, bool) {
	currency, exists := s.cache.Get(key)
	if !exists {
		return currencybus.Currency{}, false
	}

	return currency, true
}

// writeCache performs a safe write to the cache for the specified currencybus.
func (s *Store) writeCache(bus currencybus.Currency) {
	s.cache.Set(bus.ID.String(), bus)
}

// deleteCache performs a safe removal from the cache for the specified currencybus.
func (s *Store) deleteCache(bus currencybus.Currency) {
	s.cache.Delete(bus.ID.String())
}
