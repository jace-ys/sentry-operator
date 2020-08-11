package controllers

import (
	"github.com/jace-ys/sentry-operator/pkg/sentry"
)

type SentryClient struct {
	*sentry.Client
	Organization string
}

func removeFinalizer(finalizers []string, name string) []string {
	var result []string
	for _, item := range finalizers {
		if item == name {
			continue
		}

		result = append(result, item)
	}

	return result
}

func containsFinalizer(finalizers []string, name string) bool {
	for _, finalizer := range finalizers {
		if finalizer == name {
			return true
		}
	}

	return false
}
