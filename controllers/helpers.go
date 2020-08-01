package controllers

import (
	"github.com/jace-ys/sentry-operator/pkg/sentry"
)

type SentryClient struct {
	*sentry.Client
	Organization string
}
