// Package mid provides app level middleware support.
package mid

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"

	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// Encoder defines behavior that can encode a data model and provide
// the content type for that encoding.
type Encoder interface {
	Encode() (data []byte, contentType string, err error)
}

// TableInfo represents the structure and metadata of a database table
type TableInfo struct {
	Name   string                // Table name
	Action permissionsbus.Action // "create", "read", "update", "delete"
}

// isError tests if the Encoder has an error inside of it.
func isError(e Encoder) error {
	err, isError := e.(error)
	if isError {
		return err
	}
	return nil
}

// HandlerFunc represents an api layer handler function that needs to be called.
type HandlerFunc func(ctx context.Context) Encoder

// =============================================================================

type ctxKey int

const (
	claimKey ctxKey = iota + 1
	userIDKey
	userKey
	productKey
	homeKey
	trKey
	tableInfoKey
	restrictedColumnKey
)

func setClaims(ctx context.Context, claims auth.Claims) context.Context {
	return context.WithValue(ctx, claimKey, claims)
}

// GetClaims returns the claims from the context.
func GetClaims(ctx context.Context) auth.Claims {
	v, ok := ctx.Value(claimKey).(auth.Claims)
	if !ok {
		return auth.Claims{}
	}
	return v
}

func setUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID returns the user id from the context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("user id not found in context")
	}

	return v, nil
}

func setUser(ctx context.Context, usr userbus.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// GetUser returns the user from the context.
func GetUser(ctx context.Context) (userbus.User, error) {
	v, ok := ctx.Value(userKey).(userbus.User)
	if !ok {
		return userbus.User{}, errors.New("user not found in context")
	}

	return v, nil
}

func setHome(ctx context.Context, hme homebus.Home) context.Context {
	return context.WithValue(ctx, homeKey, hme)
}

// GetHome returns the home from the context.
func GetHome(ctx context.Context) (homebus.Home, error) {
	v, ok := ctx.Value(homeKey).(homebus.Home)
	if !ok {
		return homebus.Home{}, errors.New("home not found in context")
	}

	return v, nil
}

func setTran(ctx context.Context, tx sqldb.CommitRollbacker) context.Context {
	return context.WithValue(ctx, trKey, tx)
}

// GetTran retrieves the value that can manage a transaction.
func GetTran(ctx context.Context) (sqldb.CommitRollbacker, error) {
	v, ok := ctx.Value(trKey).(sqldb.CommitRollbacker)
	if !ok {
		return nil, errors.New("transaction not found in context")
	}

	return v, nil
}

func setTableInfo(ctx context.Context, tableInfo *TableInfo) context.Context {
	return context.WithValue(ctx, tableInfoKey, tableInfo)
}

func GetTableInfo(ctx context.Context) (*TableInfo, error) {
	v, ok := ctx.Value(tableInfoKey).(*TableInfo)
	if !ok {
		return nil, errors.New("table info not found in context")
	}

	return v, nil
}
