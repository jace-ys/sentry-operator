package controllers

import (
	"errors"
	"net/http"
)

var (
	ErrOutOfSync = errors.New("out of sync")
)

type retryableError struct {
	err error
}

func (e retryableError) Error() string {
	return e.err.Error()
}

func retryableHTTPError(resp *http.Response, err error) error {
	if resp.StatusCode >= 500 {
		return retryableError{err}
	}

	return err
}
