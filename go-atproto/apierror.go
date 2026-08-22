package client

import (
	"errors"
	"fmt"

	"github.com/bluesky-social/indigo/atproto/atclient"
)

// APIError is an error response returned by the PDS or proxied Gyoka service.
//
// Use errors.As to inspect its HTTP status code and AT Protocol error fields.
type APIError struct {
	StatusCode int
	Name       string
	Message    string

	cause error
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		if e.Name != "" && e.Message != "" {
			return fmt.Sprintf("API request failed (HTTP %d): %s: %s", e.StatusCode, e.Name, e.Message)
		}
		return fmt.Sprintf("API request failed (HTTP %d)", e.StatusCode)
	}
	return "API request failed"
}

// Unwrap returns the underlying Indigo error.
func (e *APIError) Unwrap() error {
	return e.cause
}

func normalizeAPIError(err error) error {
	if err == nil {
		return nil
	}

	var indigoErr *atclient.APIError
	if !errors.As(err, &indigoErr) {
		return err
	}
	return &APIError{
		StatusCode: indigoErr.StatusCode,
		Name:       indigoErr.Name,
		Message:    indigoErr.Message,
		cause:      err,
	}
}

func normalizeResult[T any](output *T, err error) (*T, error) {
	return output, normalizeAPIError(err)
}
