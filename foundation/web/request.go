package web

import (
	"fmt"
	"io"
	"net/http"
)

// Param returns the web call parameters from the request.
func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}

// Decoder represents data that can be decoded.
type Decoder interface {
	Decode(data []byte) error
}

type validator interface {
	Validate() error
}

// MaxBodyBytes caps how much of a request body Decode will read into memory.
// Without a cap, a single client can exhaust server memory by streaming an
// arbitrarily large body. 50 MiB is far above any legitimate JSON payload in
// this API while still bounding worst-case memory per request.
const MaxBodyBytes = 50 << 20

// Decode reads the body of an HTTP request and decodes the body into the
// specified data model. If the data model implements the validator interface,
// the method will be called.
func Decode(r *http.Request, v Decoder) error {
	// Read at most MaxBodyBytes+1 so an over-limit body can be detected without
	// reading it all into memory.
	data, err := io.ReadAll(io.LimitReader(r.Body, MaxBodyBytes+1))
	if err != nil {
		return fmt.Errorf("request: unable to read payload: %w", err)
	}
	if len(data) > MaxBodyBytes {
		return fmt.Errorf("request: payload exceeds %d byte limit", MaxBodyBytes)
	}

	if err := v.Decode(data); err != nil {
		return fmt.Errorf("request: decode: %w", err)
	}

	if v, ok := v.(validator); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}
