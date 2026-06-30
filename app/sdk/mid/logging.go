package mid

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// sensitiveQueryParam matches token-bearing query keys whose values must never
// reach the logs (e.g. the WebSocket connection passes the JWT as ?token=).
var sensitiveQueryParam = regexp.MustCompile(`(?i)(^|&)(token|access_token)=[^&]*`)

// scrubQuery redacts the values of sensitive query parameters so bearer tokens
// are never written to access logs. Key names and parameter order are kept;
// only the value is replaced with REDACTED.
func scrubQuery(rawQuery string) string {
	return sensitiveQueryParam.ReplaceAllString(rawQuery, "${1}${2}=REDACTED")
}

// Logger writes information about the request to the logs.
func Logger(ctx context.Context, log *logger.Logger, path string, rawQuery string, method string, remoteAddr string, next HandlerFunc) Encoder {
	now := time.Now()

	if rawQuery != "" {
		path = fmt.Sprintf("%s?%s", path, scrubQuery(rawQuery))
	}

	log.Info(ctx, "request started", "method", method, "path", path, "remoteaddr", remoteAddr)

	resp := next(ctx)
	err := isError(resp)

	var statusCode = errs.OK
	if err != nil {
		statusCode = errs.Internal

		var v *errs.Error
		if errors.As(err, &v) {
			statusCode = v.Code
		}
	}

	log.Info(ctx, "request completed", "method", method, "path", path, "remoteaddr", remoteAddr,
		"statuscode", statusCode, "since", time.Since(now).String())

	return resp
}
