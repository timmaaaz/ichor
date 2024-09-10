package mid

import (
	"context"
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/sdk/mid"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/sqldb"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
)

// BeginCommitRollback executes the transaction middleware functionality.
func BeginCommitRollback(log *logger.Logger, bgn sqldb.Beginner) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.BeginCommitRollback(ctx, log, bgn, next)
	}

	return addMidFunc(midFunc)
}
