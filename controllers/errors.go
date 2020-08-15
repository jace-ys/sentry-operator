package controllers

import (
	"errors"
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
